package tasks

import (
	"errors"
	"github.com/dokku/docket/subprocess"
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
	StorageEnsureTask StorageEnsureTask `yaml:"dokku_storage_ensure"`
}

// GetName returns the name of the example
func (e StorageEnsureTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the storage ensure task
func (t StorageEnsureTask) Doc() string {
	return "Ensures the storage for a given dokku application"
}

// Examples returns the examples for the storage ensure task
func (t StorageEnsureTask) Examples() ([]Doc, error) {
	return MarshalExamples([]StorageEnsureTaskExample{})
}

// Execute ensures the storage for a given app
func (t StorageEnsureTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return ensureStorage(t.App, t.Chown) },
		"absent":  func() TaskOutputState { return removeStorage(t.App, t.Chown) },
	})
}

// Plan reports the drift the StorageEnsureTask would produce.
//
// storage:ensure-directory does not expose a probe for whether the directory
// already exists; the plan is conservative and reports drift unconditionally.
func (t StorageEnsureTask) Plan() PlanResult {
	chownValues := map[string]bool{
		"heroku": true, "herokuish": true, "paketo": true, "root": true, "false": true,
	}
	return DispatchPlan(t.State, map[State]func() PlanResult{
		"present": func() PlanResult {
			if !chownValues[t.Chown] {
				return PlanResult{Status: PlanStatusError, Error: errors.New("invalid chown value specified")}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    "directory presence not probed",
				Mutations: []string{"storage:ensure-directory --chown " + t.Chown + " " + t.App},
			}
		},
		"absent": func() PlanResult {
			return PlanResult{Status: PlanStatusError, Error: errors.New("the absent state is not supported for storage:ensure")}
		},
	})
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
		return TaskOutputErrorFromExec(state, err, result)
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
