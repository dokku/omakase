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
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the StorageEnsureTask would produce. dokku does
// not expose a probe for storage:ensure-directory, so the plan reports
// drift unconditionally.
func (t StorageEnsureTask) Plan() PlanResult {
	chownValues := map[string]bool{
		"heroku": true, "herokuish": true, "paketo": true, "root": true, "false": true,
	}
	return DispatchPlan(t.State, map[State]func() PlanResult{
		StatePresent: func() PlanResult {
			if !chownValues[t.Chown] {
				return PlanResult{Status: PlanStatusError, Error: errors.New("invalid chown value specified")}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    "directory presence not probed",
				Mutations: []string{"storage:ensure-directory --chown " + t.Chown + " " + t.App},
				apply: func() TaskOutputState {
					state := TaskOutputState{Changed: false, State: StateAbsent}
					result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
						Command: "dokku",
						Args:    []string{"--quiet", "storage:ensure-directory", "--chown", t.Chown, t.App},
					})
					if err != nil {
						return TaskOutputErrorFromExec(state, err, result)
					}
					state.Changed = true
					state.State = StatePresent
					return state
				},
			}
		},
		StateAbsent: func() PlanResult {
			return PlanResult{Status: PlanStatusError, Error: errors.New("the absent state is not supported for storage:ensure")}
		},
	})
}

// init registers the StorageEnsureTask with the task registry
func init() {
	RegisterTask(&StorageEnsureTask{})
}
