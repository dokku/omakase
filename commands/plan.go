package commands

import (
	"fmt"
	"os"

	"github.com/dokku/docket/subprocess"
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

	tasksFile         string
	host              string
	sudo              bool
	acceptNewHostKeys bool
	tags              []string
	skipTags          []string
	arguments         map[string]*Argument
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
	f.StringVar(&c.host, "host", "", "remote dokku host as [user@]host[:port]; equivalent to DOKKU_HOST. Routes every dokku invocation through ssh.")
	f.BoolVar(&c.sudo, "sudo", false, "wrap remote dokku invocations with `sudo -n`; equivalent to DOKKU_SUDO=1")
	f.BoolVar(&c.acceptNewHostKeys, "accept-new-host-keys", false, "for SSH transport, accept new host keys on first connection (`-o StrictHostKeyChecking=accept-new`). MITM risk on first connect.")
	f.StringSliceVar(&c.tags, "tags", nil, "comma-separated tag list; only tasks whose `tags:` set intersects this list are planned")
	f.StringSliceVar(&c.skipTags, "skip-tags", nil, "comma-separated tag list; tasks whose `tags:` set intersects this list are skipped")

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
			"--tasks":                complete.PredictFiles("*.yml"),
			"--host":                 complete.PredictAnything,
			"--sudo":                 complete.PredictNothing,
			"--accept-new-host-keys": complete.PredictNothing,
			"--tags":                 complete.PredictAnything,
			"--skip-tags":            complete.PredictAnything,
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

	resolvedHost := resolveSshFlags(c.host, c.sudo, c.acceptNewHostKeys)

	data, err := os.ReadFile(c.tasksFile)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("read error: %v", err))
		return 1
	}

	context := make(map[string]interface{})
	var sensitiveValues []string
	for name, argument := range c.arguments {
		if argument.Required && !argument.HasValue() {
			c.Ui.Error(fmt.Sprintf("Missing flag '--%s'", name))
			return 1
		}
		context[name] = argument.GetValue()
		if argument.Sensitive {
			if v := argument.StringValue(); v != "" {
				sensitiveValues = append(sensitiveValues, v)
			}
		}
	}

	taskList, err := tasks.GetTasks(data, context)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("task error: %v", err))
		return 1
	}

	sensitiveValues = append(sensitiveValues, tasks.CollectSensitiveValues(taskList)...)
	subprocess.SetGlobalSensitive(sensitiveValues)
	defer subprocess.SetGlobalSensitive(nil)

	if resolvedHost != "" {
		defer subprocess.CloseSshControlMaster(resolvedHost)
	}

	formatter := NewFormatter(c.Ui, false)
	formatter.PlayHeaderWithHost("tasks", resolvedHost)

	counts := PlanCounts{}
	hasError := false

	keys := tasks.FilterByTags(taskList, c.tags, c.skipTags)
	exprBaseCtx := buildEnvelopeExprContext(context)

	for _, name := range keys {
		env := taskList.GetEnvelope(name)

		if env.HasWhen() {
			ok, err := tasks.EvalBool(env.WhenProgram(), envelopeExprContext(exprBaseCtx, env))
			if err != nil {
				counts.Tasks++
				counts.Errors++
				hasError = true
				formatter.TaskLine(MarkerProbeError, name, fmt.Sprintf("(when expression error: %v)", err))
				continue
			}
			if !ok {
				counts.Tasks++
				counts.Skipped++
				formatter.TaskLine(MarkerSkipped, name, "(when: false)")
				continue
			}
		}

		result := env.Task.Plan()
		counts.Tasks++

		switch {
		case result.Error != nil:
			counts.Errors++
			hasError = true
			formatter.TaskLine(MarkerProbeError, name, fmt.Sprintf("(%s)", PrefixErrorMessage(result.Error)))
		case result.InSync:
			counts.InSync++
			formatter.TaskLine(MarkerOK, name, "(in sync)")
		default:
			counts.WouldChange++
			marker := Marker(result.Status)
			if marker == "" {
				marker = Marker(tasks.PlanStatusModify)
			}
			suffix := ""
			if result.Reason != "" {
				suffix = "(" + result.Reason + ")"
			}
			formatter.TaskLine(marker, name, suffix)
			for _, m := range result.Mutations {
				formatter.Continuation('-', m)
			}
		}
	}

	formatter.PlanSummary(counts)

	if hasError {
		return 1
	}
	return 0
}
