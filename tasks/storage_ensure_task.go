package tasks

import (
	"errors"
	"fmt"
	"omakase/subprocess"

	yaml "gopkg.in/yaml.v3"
)

// StorageEnsureTask manages the storage for a given dokku application
type StorageEnsureTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Chown is the chown value to set
	Chown string `required:"false" yaml:"chown"`

	// State is the desired state of the storage
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// StorageEnsureTaskExample contains an example of a StorageEnsureTask
type StorageEnsureTaskExample struct {
	// Name is the task name holding the StorageEnsureTask description
	Name string `yaml:"-"`

	// StorageEnsureTask is the StorageEnsureTask configuration
	StorageEnsureTask StorageEnsureTask `yaml:"storage_ensure"`
}

// DesiredState returns the desired state of the storage
func (t StorageEnsureTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the storage ensure task
func (t StorageEnsureTask) Doc() string {
	return "Ensures the storage for a given dokku application"
}

// Examples returns the examples for the builder property task
func (t StorageEnsureTask) Examples() ([]Doc, error) {
	examples := []StorageEnsureTaskExample{}

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

// Execute ensures the storage for a given app
func (t StorageEnsureTask) Execute() TaskOutputState {
	funcMap := map[State]func(string, string) TaskOutputState{
		"present": ensureStorage,
		"absent":  removeStorage,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
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
		"paketo":    true,
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
