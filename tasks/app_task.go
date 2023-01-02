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
		"present": func(app string) TaskOutputState {
			state := TaskOutputState{
				Changed: false,
				State:   "absent",
			}
			if appExists(t.App) {
				state.State = "present"
				return state
			}

			resp := runDokkuCommand([]string{"--quiet", "apps:create", t.App})
			if resp.HasError() {
				state.Error = resp.Error
				state.Message = string(resp.Stderr)
				return state
			}

			state.Changed = true
			state.State = "present"
			return state
		},
		"absent": func(app string) TaskOutputState {
			state := TaskOutputState{
				Changed: false,
				State:   "present",
			}
			if !appExists(t.App) {
				state.State = "absent"
				return state
			}

			resp := runDokkuCommand([]string{"--quiet", "--force", "apps:destroy", t.App})
			if resp.HasError() {
				state.Error = resp.Error
				state.Message = string(resp.Stderr)
				return state
			}

			state.Changed = true
			state.State = "absent"
			return state
		},
	}

	fn := funcMap[t.State]
	return fn(t.App)
}

func appExists(appName string) bool {
	cmd := subprocess.NewShellCmdWithArgs("dokku", "--quiet", "apps:exists", appName)
	return cmd.ExecuteQuiet()
}

func init() {
	RegisterTask(&AppTask{})
}
