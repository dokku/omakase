package tasks

import (
	"fmt"
	"docket/subprocess"
)

// ToggleContext represents the context for a toggle operation
type ToggleContext struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the toggle should be applied globally
	Global bool `required:"false" yaml:"global"`

	// AllowGlobal is a flag indicating if the toggle should be applied globally
	AllowGlobal bool `required:"false" yaml:"allow_global"`
}

// enablePlugin executes the enable state for a plugin
func enablePlugin(subcommand string, pctx ToggleContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	appName := pctx.App
	if pctx.AllowGlobal {
		if pctx.Global && pctx.App != "" {
			state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
			return state
		}
		if pctx.Global {
			appName = "--global"
		}
	}

	// todo: validate that the plugin isn't already enabled
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			subcommand,
			appName,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// executeToggle is a shared Execute implementation for toggle tasks.
func executeToggle(state State, app string, global bool, allowGlobal bool, enableCmd, disableCmd string) TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: allowGlobal,
		App:         app,
		Global:      global,
	}
	return DispatchState(state, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return enablePlugin(enableCmd, ctx) },
		"absent":  func() TaskOutputState { return disablePlugin(disableCmd, ctx) },
	})
}

// disablePlugin executes the disable state for a plugin
func disablePlugin(subcommand string, pctx ToggleContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	appName := pctx.App
	if pctx.AllowGlobal {
		if pctx.Global && pctx.App != "" {
			state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
			return state
		}
		if pctx.Global {
			appName = "--global"
		}
	}

	// todo: validate that the plugin isn't already disabled
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			subcommand,
			appName,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}
