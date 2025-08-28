package tasks

import (
	"errors"
	"omakase/subprocess"
)

// StorageEnsureTask manages the storage for a given dokku application
type StorageEnsureTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Chown is the chown value to set
	Chown string `required:"false" yaml:"chown"`

	// State is the desired state of the storage
	State State `required:"true" yaml:"state" default:"present"`
}

// DesiredState returns the desired state of the storage
func (t StorageEnsureTask) DesiredState() State {
	return t.State
}

// Execute ensures the storage for a given app
func (t StorageEnsureTask) Execute() TaskOutputState {
	funcMap := map[State]func(string, string) TaskOutputState{
		"present": ensureStorage,
		"absent":  removeStorage,
	}

	fn := funcMap[t.State]
	return fn(t.App, t.Chown)
}

// ensureStorage ensures the storage for a given app
func ensureStorage(app, chown string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	// ensure chown is a valid value
	chownValues := map[string]bool{
		"heroku":    true,
		"herokuish": true,
		"packeto":   true,
		"root":      true,
		"false":     true,
	}
	if !chownValues[chown] {
		state.Error = errors.New("invalid chown value specified")
		return state
	}

	// todo: implement a check to see if the folder exists?
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"storage:ensure-directory",
			"--chown",
			chown,
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

// removeStorage removes the storage for a given app
func removeStorage(app, chown string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	state.Error = errors.New("the absent state is not supported for storage:ensure")
	return state
}

// init registers the StorageEnsureTask with the task registry
func init() {
	RegisterTask(&StorageEnsureTask{})
}
