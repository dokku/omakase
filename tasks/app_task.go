package tasks

import (
	"fmt"
	"docket/subprocess"

	yaml "gopkg.in/yaml.v3"
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

// DesiredState returns the desired state of the app
func (t AppTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the app task
func (t AppTask) Doc() string {
	return "Creates or destroys an app"
}

// Examples returns a list of AppTaskExamples as yaml
func (t AppTask) Examples() ([]Doc, error) {
	examples := []AppTaskExample{
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
	}

	var output []Doc
	for _, example := range examples {
		b, err := yaml.Marshal(example)
		if err != nil {
			return nil, err
		}

		output = append(output, Doc{
			Name:      example.Name,
			Codeblock: string(b),
		})
	}

	return output, nil
}

// Execute creates or destroys an app
func (t AppTask) Execute() TaskOutputState {
	funcMap := map[State]func(string) TaskOutputState{
		"present": createApp,
		"absent":  destroyApp,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
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
