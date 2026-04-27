package commands

import (
	"fmt"
	"os"

	"github.com/dokku/docket/tasks"

	"github.com/josegonzalez/cli-skeleton/command"
	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
)

// PlanCommand reports the drift each task in a docket recipe would produce
// against the live server, without mutating it. Plan is fully driven by the
// per-task Plan() method; the apply path is never invoked.
type PlanCommand struct {
	command.Meta

	tasksFile string
	arguments map[string]*Argument
}

func (c *PlanCommand) Name() string {
	return "plan"
}

func (c *PlanCommand) Synopsis() string {
	return "Reports the drift a docket task file would produce, without mutating state"
}

func (c *PlanCommand) Help() string {
	return command.CommandHelp(c)
}

func (c *PlanCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"Plan tasks from the default tasks.yml": fmt.Sprintf("%s %s", appName, c.Name()),
		"Plan tasks from a specific file":       fmt.Sprintf("%s %s --tasks path/to/task.yml", appName, c.Name()),
		"Plan tasks from a remote URL":          fmt.Sprintf("%s %s --tasks http://dokku.com/docket/example.yml", appName, c.Name()),
		"Override a task input":                 fmt.Sprintf("%s %s --name lollipop", appName, c.Name()),
	}
}

func (c *PlanCommand) Arguments() []command.Argument {
	return []command.Argument{}
}

func (c *PlanCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *PlanCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *PlanCommand) FlagSet() *flag.FlagSet {
	f := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	f.StringVar(&c.tasksFile, "tasks", "tasks.yml", "a yaml file containing a task list")

	taskFile := getTaskYamlFilename(os.Args)
	data, err := os.ReadFile(taskFile)
	if err != nil {
		return f
	}

	arguments, err := registerInputFlags(f, data)
	if err != nil {
		return f
	}
	c.arguments = arguments

	return f
}

func (c *PlanCommand) AutocompleteFlags() complete.Flags {
	return command.MergeAutocompleteFlags(
		c.Meta.AutocompleteFlags(command.FlagSetClient),
		complete.Flags{
			"--tasks": complete.PredictFiles("*.yml"),
		},
	)
}

// Run iterates every task in the parsed recipe, invokes Plan() (read-only
// by contract), and prints a one-line summary per task plus a final
// summary line.
//
// Exit codes:
//
//	0 - plan completed successfully (regardless of drift)
//	1 - read error, parse error, or read-state probe error
func (c *PlanCommand) Run(args []string) int {
	flags := c.FlagSet()
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		c.Ui.Error(err.Error())
		c.Ui.Error(command.CommandErrorText(c))
		return 1
	}

	data, err := os.ReadFile(c.tasksFile)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("read error: %v", err))
		return 1
	}

	context := make(map[string]interface{})
	for name, argument := range c.arguments {
		if argument.Required && !argument.HasValue() {
			c.Ui.Error(fmt.Sprintf("Missing flag '--%s'", name))
			return 1
		}
		context[name] = argument.GetValue()
	}

	taskList, err := tasks.GetTasks(data, context)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("task error: %v", err))
		return 1
	}

	totals := struct {
		tasks       int
		wouldChange int
		inSync      int
		errors      int
	}{}
	hasError := false

	for _, name := range taskList.Keys() {
		task := taskList.Get(name)
		result := task.Plan()
		totals.tasks++

		switch {
		case result.Error != nil:
			totals.errors++
			hasError = true
			c.Ui.Error(fmt.Sprintf("[error]   %s  (%v)", name, result.Error))
		case result.InSync:
			totals.inSync++
			c.Ui.Info(fmt.Sprintf("[ok]      %s  (in sync)", name))
		default:
			totals.wouldChange++
			marker := string(result.Status)
			if marker == "" {
				marker = string(tasks.PlanStatusModify)
			}
			line := fmt.Sprintf("[%s]       %s", marker, name)
			if result.Reason != "" {
				line = fmt.Sprintf("[%s]       %s  (%s)", marker, name, result.Reason)
			}
			c.Ui.Info(line)
			for _, m := range result.Mutations {
				c.Ui.Info(fmt.Sprintf("          - %s", m))
			}
		}
	}

	c.Ui.Info("")
	c.Ui.Info(fmt.Sprintf(
		"Plan: %d task(s); %d would change, %d in sync, %d error(s).",
		totals.tasks, totals.wouldChange, totals.inSync, totals.errors,
	))

	if hasError {
		return 1
	}
	return 0
}
