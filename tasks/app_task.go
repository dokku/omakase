package tasks

import (
	"omakase/subprocess"
)

// AppTask creates or destroys an app
type AppTask struct {
	// App is the name of the app
	App   string `required:"true" yaml:"app"`

	// State is the state of the app
	State State `required:"true" yaml:"state" default:"present" options:"present,absent"`
}

}

// DesiredState returns the desired state of the app
func (t AppTask) DesiredState() State {
	return t.State
}

// Execute creates or destroys an app
func (t AppTask) Execute() TaskOutputState {
	funcMap := map[State]func(string) TaskOutputState{
		"present": createApp,
		"absent":  destroyApp,
	}

	fn := funcMap[t.State]
	return fn(t.App)
}

// appExists checks if an app exists
func appExists(appName string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"apps:exists",
			appName,
		},
	})
	if err != nil {
		return false
	}

	return result.ExitCode == 0
}

// createApp creates an app
func createApp(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}
	if appExists(app) {
		state.State = "present"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"apps:create",
			app,
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

// destroyApp destroys an app
func destroyApp(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	if !appExists(app) {
		state.State = "absent"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"--force",
			"apps:destroy",
			app,
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

// init registers the AppTask with the task registry
func init() {
	RegisterTask(&AppTask{})
}
