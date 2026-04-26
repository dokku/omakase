package tasks

import (
	"fmt"

	"github.com/dokku/docket/subprocess"
)

// AppCloneTask clones an existing dokku app to a new app
type AppCloneTask struct {
	// App is the name of the new (target) app
	App string `required:"true" yaml:"app"`

	// SourceApp is the name of the existing app to clone from
	SourceApp string `required:"true" yaml:"source_app"`

	// SkipDeploy skips deployment of the cloned app
	SkipDeploy bool `required:"false" yaml:"skip_deploy,omitempty"`

	// State is the desired state of the cloned app
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present"`
}

// AppCloneTaskExample contains an example of an AppCloneTask
type AppCloneTaskExample struct {
	// Name is the task name holding the AppCloneTask description
	Name string `yaml:"-"`

	// AppCloneTask is the AppCloneTask configuration
	AppCloneTask AppCloneTask `yaml:"dokku_app_clone"`
}

// GetName returns the name of the example
func (e AppCloneTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the app clone task
func (t AppCloneTask) Doc() string {
	return "Clones an existing dokku app to a new app"
}

// Examples returns the examples for the app clone task
func (t AppCloneTask) Examples() ([]Doc, error) {
	return MarshalExamples([]AppCloneTaskExample{
		{
			Name: "Clone an app",
			AppCloneTask: AppCloneTask{
				App:       "node-js-app-staging",
				SourceApp: "node-js-app",
			},
		},
		{
			Name: "Clone an app without deploying",
			AppCloneTask: AppCloneTask{
				App:        "node-js-app-staging",
				SourceApp:  "node-js-app",
				SkipDeploy: true,
			},
		},
	})
}

// Execute clones an existing dokku app to a new app
func (t AppCloneTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		StatePresent: func() TaskOutputState { return cloneApp(t) },
	})
}

// cloneApp clones an existing dokku app to a new app
func cloneApp(t AppCloneTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if t.App == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}
	if t.SourceApp == "" {
		state.Error = fmt.Errorf("'source_app' is required")
		return state
	}

	if appExists(t.App) {
		state.State = StatePresent
		return state
	}

	args := []string{"--quiet", "apps:clone"}
	if t.SkipDeploy {
		args = append(args, "--skip-deploy")
	}
	args = append(args, t.SourceApp, t.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StatePresent
	return state
}

// init registers the AppCloneTask with the task registry
func init() {
	RegisterTask(&AppCloneTask{})
}
