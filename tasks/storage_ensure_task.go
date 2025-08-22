package tasks

import (
	"errors"
	"omakase/subprocess"
)

type StorageEnsureTask struct {
	App   string `required:"true" yaml:"app"`
	Chown string `required:"false" yaml:"chown"`
	State string `required:"true" yaml:"state" default:"present"`
}

func (t StorageEnsureTask) DesiredState() string {
	return t.State
}

func (t StorageEnsureTask) Execute() TaskOutputState {
	funcMap := map[string]func(string, string) TaskOutputState{
		"present": ensureStorage,
		"absent":  removeStorage,
	}

	fn := funcMap[t.State]
	return fn(t.App, t.Chown)
}

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

func removeStorage(app, chown string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	state.Error = errors.New("the absent state is not supported for storage:ensure")
	return state
}

func init() {
	RegisterTask(&StorageEnsureTask{})
}
