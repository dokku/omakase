package tasks

import (
	"fmt"
	"omakase/subprocess"
)

type ToggleContext struct {
	App         string `required:"true" yaml:"app"`
	Global      bool   `required:"false" yaml:"global"`
	AllowGlobal bool   `required:"false" yaml:"allow_global"`
}

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
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "present"
	return state
}

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
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}
