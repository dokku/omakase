package tasks

import (
	"fmt"

	"github.com/dokku/docket/subprocess"
)

// AppLockTask locks or unlocks a dokku application from deployment
type AppLockTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// State is the desired lock state
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// AppLockTaskExample contains an example of an AppLockTask
type AppLockTaskExample struct {
	// Name is the task name holding the AppLockTask description
	Name string `yaml:"-"`

	// AppLockTask is the AppLockTask configuration
	AppLockTask AppLockTask `yaml:"dokku_app_lock"`
}

// GetName returns the name of the example
func (e AppLockTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the app lock task
func (t AppLockTask) Doc() string {
	return "Locks or unlocks a dokku application from deployment"
}

// Examples returns the examples for the app lock task
func (t AppLockTask) Examples() ([]Doc, error) {
	return MarshalExamples([]AppLockTaskExample{
		{
			Name: "Lock an app",
			AppLockTask: AppLockTask{
				App: "node-js-app",
			},
		},
		{
			Name: "Unlock an app",
			AppLockTask: AppLockTask{
				App:   "node-js-app",
				State: StateAbsent,
			},
		},
	})
}

// Execute locks or unlocks the app
func (t AppLockTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		StatePresent: func() TaskOutputState { return lockApp(t.App) },
		StateAbsent:  func() TaskOutputState { return unlockApp(t.App) },
	})
}

// Plan reports the drift the AppLockTask would produce.
func (t AppLockTask) Plan() PlanResult {
	if t.App == "" {
		return PlanResult{Status: PlanStatusError, Error: fmt.Errorf("'app' is required")}
	}
	return DispatchPlan(t.State, map[State]func() PlanResult{
		StatePresent: func() PlanResult {
			if appLocked(t.App) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    "app unlocked",
				Mutations: []string{"lock " + t.App},
			}
		},
		StateAbsent: func() PlanResult {
			if !appLocked(t.App) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    "app locked",
				Mutations: []string{"unlock " + t.App},
			}
		},
	})
}

// appLocked checks if a dokku app is locked
func appLocked(app string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "apps:locked", app},
	})
	if err != nil {
		return false
	}
	return result.ExitCode == 0
}

// lockApp locks a dokku app for deployment
func lockApp(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if app == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}

	if appLocked(app) {
		state.State = StatePresent
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "apps:lock", app},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StatePresent
	return state
}

// unlockApp unlocks a dokku app for deployment
func unlockApp(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StatePresent,
	}

	if app == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}

	if !appLocked(app) {
		state.State = StateAbsent
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "apps:unlock", app},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StateAbsent
	return state
}

// init registers the AppLockTask with the task registry
func init() {
	RegisterTask(&AppLockTask{})
}
