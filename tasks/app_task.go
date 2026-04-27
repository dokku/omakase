package tasks

import (
	"github.com/dokku/docket/subprocess"
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
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the AppTask would produce.
func (t AppTask) Plan() PlanResult {
	return DispatchPlan(t.State, map[State]func() PlanResult{
		StatePresent: func() PlanResult {
			if appExists(t.App) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusCreate,
				Reason:    "app missing",
				Mutations: []string{"create app " + t.App},
				apply:     applyCreateApp(t.App),
			}
		},
		StateAbsent: func() PlanResult {
			if !appExists(t.App) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusDestroy,
				Reason:    "app present",
				Mutations: []string{"destroy app " + t.App},
				apply:     applyDestroyApp(t.App),
			}
		},
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

// applyCreateApp returns a closure that runs `dokku apps:create <app>`.
func applyCreateApp(app string) func() TaskOutputState {
	return func() TaskOutputState {
		state := TaskOutputState{Changed: false, State: StateAbsent}
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "apps:create", app},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
		state.Changed = true
		state.State = StatePresent
		return state
	}
}

// applyDestroyApp returns a closure that runs `dokku --force apps:destroy <app>`.
func applyDestroyApp(app string) func() TaskOutputState {
	return func() TaskOutputState {
		state := TaskOutputState{Changed: false, State: StatePresent}
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "--force", "apps:destroy", app},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
		state.Changed = true
		state.State = StateAbsent
		return state
	}
}

// destroyApp is retained as an integration-test helper. It runs the
// destroy-app apply path synchronously.
func destroyApp(app string) TaskOutputState {
	if !appExists(app) {
		return TaskOutputState{Changed: false, State: StateAbsent}
	}
	return applyDestroyApp(app)()
}

// createApp is retained as an integration-test helper. It runs the
// create-app apply path synchronously.
func createApp(app string) TaskOutputState {
	if appExists(app) {
		return TaskOutputState{Changed: false, State: StatePresent}
	}
	return applyCreateApp(app)()
}

// init registers the AppTask with the task registry
func init() {
	RegisterTask(&AppTask{})
}
