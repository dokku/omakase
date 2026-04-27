package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/dokku/docket/tasks"

	"github.com/josegonzalez/cli-skeleton/command"
	"github.com/posener/complete"
	flag "github.com/spf13/pflag"
)

type ApplyCommand struct {
	command.Meta

	tasksFile string
	verbose   bool
	arguments map[string]*Argument
}

func (c *ApplyCommand) Name() string {
	return "apply"
}

func (c *ApplyCommand) Synopsis() string {
	return "Applies a docket task file"
}

func (c *ApplyCommand) Help() string {
	return command.CommandHelp(c)
}

func (c *ApplyCommand) Examples() map[string]string {
	appName := os.Getenv("CLI_APP_NAME")
	return map[string]string{
		"Apply tasks from the default tasks.yml": fmt.Sprintf("%s %s", appName, c.Name()),
		"Apply tasks from a specific file":       fmt.Sprintf("%s %s --tasks path/to/task.yml", appName, c.Name()),
		"Apply tasks from a remote URL":          fmt.Sprintf("%s %s --tasks http://dokku.com/docket/example.yml", appName, c.Name()),
		"Override a task input":                  fmt.Sprintf("%s %s --name lollipop", appName, c.Name()),
	}
}

func (c *ApplyCommand) Arguments() []command.Argument {
	return []command.Argument{}
}

func (c *ApplyCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *ApplyCommand) ParsedArguments(args []string) (map[string]command.Argument, error) {
	return command.ParseArguments(args, c.Arguments())
}

func (c *ApplyCommand) FlagSet() *flag.FlagSet {
	f := c.Meta.FlagSet(c.Name(), command.FlagSetClient)
	f.StringVar(&c.tasksFile, "tasks", "tasks.yml", "a yaml file containing a task list")
	f.BoolVar(&c.verbose, "verbose", false, "echo the resolved dokku command for each task as a continuation line (commands are not masked; avoid on recipes that pass secrets via task arguments)")

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

func (c *ApplyCommand) AutocompleteFlags() complete.Flags {
	return command.MergeAutocompleteFlags(
		c.Meta.AutocompleteFlags(command.FlagSetClient),
		complete.Flags{
			"--tasks":   complete.PredictFiles("*.yml"),
			"--verbose": complete.PredictNothing,
		},
	)
}

func (c *ApplyCommand) Run(args []string) int {
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

	formatter := NewFormatter(c.Ui, c.verbose)
	formatter.PlayHeader("tasks")

	start := time.Now()
	counts := ApplyCounts{}

	for _, name := range taskList.Keys() {
		task := taskList.Get(name)
		state := task.Execute()
		counts.Tasks++

		switch {
		case state.Error != nil:
			counts.Errors++
			formatter.TaskLine(MarkerError, name, "")
			formatter.Continuation('!', state.Error.Error())
			if c.verbose {
				for _, cmd := range state.Commands {
					formatter.Continuation('\u2192', cmd)
				}
			}
			formatter.ApplySummary(counts, time.Since(start))
			return 1
		case state.Changed:
			counts.Changed++
			formatter.TaskLine(MarkerChanged, name, "")
		default:
			counts.OK++
			formatter.TaskLine(MarkerOK, name, "")
		}

		if c.verbose {
			for _, cmd := range state.Commands {
				formatter.Continuation('\u2192', cmd)
			}
		}

		if state.State != state.DesiredState {
			counts.Errors++
			formatter.TaskLine(MarkerError, name, "")
			formatter.Continuation('!', fmt.Sprintf("invalid state: expected=%v actual=%v", state.DesiredState, state.State))
			formatter.ApplySummary(counts, time.Since(start))
			return 1
		}
	}

	formatter.ApplySummary(counts, time.Since(start))
	return 0
}
