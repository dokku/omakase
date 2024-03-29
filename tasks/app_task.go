package tasks

import (
	"omakase/subprocess"
)

type AppTask struct {
	App   string `required:"true" yaml:"app"`
	State string `required:"true" yaml:"state" default:"present"`
}

func (t AppTask) DesiredState() string {
	return t.State
}

func (t AppTask) Execute() TaskOutputState {
	funcMap := map[string]func(string) TaskOutputState{
		"present": createApp,
		"absent":  destroyApp,
	}

	fn := funcMap[t.State]
	return fn(t.App)
}

func appExists(appName string) bool {
	cmd := subprocess.NewShellCmdWithArgs("dokku", "--quiet", "apps:exists", appName)
	return cmd.ExecuteQuiet()
}

func createApp(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}
	if appExists(app) {
		state.State = "present"
		return state
	}

	resp := subprocess.RunDokkuCommand([]string{"--quiet", "apps:create", app})
	if resp.HasError() {
		state.Error = resp.Error
		state.Message = string(resp.Stderr)
		return state
	}

	state.Changed = true
	state.State = "present"
	return state
}

func destroyApp(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	if !appExists(app) {
		state.State = "absent"
		return state
	}

	resp := subprocess.RunDokkuCommand([]string{"--quiet", "--force", "apps:destroy", app})
	if resp.HasError() {
		state.Error = resp.Error
		state.Message = string(resp.Stderr)
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}

func init() {
	RegisterTask(&AppTask{})
}
