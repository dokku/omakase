package tasks

import (
	"docket/subprocess"
)

// AppTask creates or destroys an app
type AppTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// State is the state of the app
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// AppTaskExample contains an example of an AppTask
type AppTaskExample struct {
	// Name is the task name holding the AppTask description
	Name string `yaml:"-"`

	// DokkuApp is the AppTask configuration
	DokkuApp AppTask `yaml:"dokku_app"`
}

// GetName returns the name of the example
func (e AppTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the app task
func (t AppTask) Doc() string {
	return "Creates or destroys an app"
}

// Examples returns a list of AppTaskExamples as yaml
func (t AppTask) Examples() ([]Doc, error) {
	return MarshalExamples([]AppTaskExample{
		{
			Name: "Create an app named hello-world",
			DokkuApp: AppTask{
				App: "hello-world",
			},
		},
		{
			Name: "Destroy the app named hello-world",
			DokkuApp: AppTask{
				App:   "hello-world",
				State: "absent",
			},
		},
	})
}

// Execute creates or destroys an app
func (t AppTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return createApp(t.App) },
		"absent":  func() TaskOutputState { return destroyApp(t.App) },
	})
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
		return TaskOutputErrorFromExec(state, err, result)
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
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the AppTask with the task registry
func init() {
	RegisterTask(&AppTask{})
}
