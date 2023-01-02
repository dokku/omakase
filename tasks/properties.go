package tasks

import (
	"fmt"
	"omakase/subprocess"
)

type PropertyContext struct {
	App      string `required:"true" yaml:"app"`
	Global   bool   `required:"false" yaml:"global"`
	Property string `required:"true" yaml:"property"`
	Value    string `required:"false" yaml:"value"`
}

func setProperty(subcommand string, pctx PropertyContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	if pctx.Global && pctx.App != "" {
		state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
		return state
	}

	appName := "--global"
	if pctx.App != "" {
		appName = pctx.App
	}

	if pctx.Value == "" {
		state.Error = fmt.Errorf("setting a state of 'present' is invalid without a value for 'value'")
		return state
	}

	// todo: validate that the value isn't already set

	resp := subprocess.RunDokkuCommand([]string{"--quiet", subcommand, appName, pctx.Property, pctx.Value})
	if resp.HasError() {
		state.Error = resp.Error
		state.Message = string(resp.Stderr)
		return state
	}

	state.Changed = true
	state.State = "present"
	return state
}

func unsetProperty(subcommand string, pctx PropertyContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	if pctx.Global && pctx.App != "" {
		state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
		return state
	}

	appName := "--global"
	if pctx.App != "" {
		appName = pctx.App
	}

	if pctx.Value != "" {
		state.Error = fmt.Errorf("setting a state of 'absent' is invalid with a value for 'value'")
		return state
	}

	// todo: validate that the value isn't already unset

	resp := subprocess.RunDokkuCommand([]string{"--quiet", subcommand, appName, pctx.Property})
	if resp.HasError() {
		state.Error = resp.Error
		state.Message = string(resp.Stderr)
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}
