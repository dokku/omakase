package tasks

import (
	"fmt"
	"strings"

	"github.com/dokku/docket/subprocess"
)

// BuildpacksTask manages the buildpacks for a given dokku application
type BuildpacksTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Buildpacks is the list of buildpack URLs
	Buildpacks []string `required:"false" yaml:"buildpacks"`

	// State is the desired state of the buildpacks
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// BuildpacksTaskExample contains an example of a BuildpacksTask
type BuildpacksTaskExample struct {
	// Name is the task name holding the BuildpacksTask description
	Name string `yaml:"-"`

	// BuildpacksTask is the BuildpacksTask configuration
	BuildpacksTask BuildpacksTask `yaml:"dokku_buildpacks"`
}

// GetName returns the name of the example
func (e BuildpacksTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the buildpacks task
func (t BuildpacksTask) Doc() string {
	return "Manages the buildpacks for a given dokku application"
}

// Examples returns the examples for the buildpacks task
func (t BuildpacksTask) Examples() ([]Doc, error) {
	return MarshalExamples([]BuildpacksTaskExample{
		{
			Name: "Add buildpacks to an app",
			BuildpacksTask: BuildpacksTask{
				App: "node-js-app",
				Buildpacks: []string{
					"https://github.com/heroku/heroku-buildpack-nodejs.git",
					"https://github.com/heroku/heroku-buildpack-nginx.git",
				},
			},
		},
		{
			Name: "Remove a buildpack from an app",
			BuildpacksTask: BuildpacksTask{
				App: "node-js-app",
				Buildpacks: []string{
					"https://github.com/heroku/heroku-buildpack-nginx.git",
				},
				State: StateAbsent,
			},
		},
		{
			Name: "Clear all buildpacks from an app",
			BuildpacksTask: BuildpacksTask{
				App:   "node-js-app",
				State: StateAbsent,
			},
		},
	})
}

// Execute manages the buildpacks for a given app
func (t BuildpacksTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		StatePresent: func() TaskOutputState { return addBuildpacks(t) },
		StateAbsent:  func() TaskOutputState { return removeBuildpacks(t) },
	})
}

// getBuildpacks fetches the current buildpacks list for an app
func getBuildpacks(app string) (map[string]bool, error) {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "buildpacks:list", app},
	})
	if err != nil {
		return nil, err
	}

	buildpacks := map[string]bool{}
	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "=====>") {
			continue
		}
		buildpacks[trimmed] = true
	}
	return buildpacks, nil
}

// addBuildpacks adds buildpacks to an app, skipping ones already present
func addBuildpacks(t BuildpacksTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if t.App == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}
	if len(t.Buildpacks) == 0 {
		state.Error = fmt.Errorf("'buildpacks' must not be empty for state 'present'")
		return state
	}

	current, err := getBuildpacks(t.App)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	var toAdd []string
	for _, bp := range t.Buildpacks {
		if !current[bp] {
			toAdd = append(toAdd, bp)
		}
	}

	if len(toAdd) == 0 {
		state.State = StatePresent
		return state
	}

	for _, bp := range toAdd {
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "buildpacks:add", t.App, bp},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
	}

	state.Changed = true
	state.State = StatePresent
	return state
}

// removeBuildpacks removes buildpacks from an app, or clears all when none are specified
func removeBuildpacks(t BuildpacksTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StatePresent,
	}

	if t.App == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}

	current, err := getBuildpacks(t.App)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	if len(t.Buildpacks) == 0 {
		if len(current) == 0 {
			state.State = StateAbsent
			return state
		}
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "buildpacks:clear", t.App},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
		state.Changed = true
		state.State = StateAbsent
		return state
	}

	var toRemove []string
	for _, bp := range t.Buildpacks {
		if current[bp] {
			toRemove = append(toRemove, bp)
		}
	}

	if len(toRemove) == 0 {
		state.State = StateAbsent
		return state
	}

	for _, bp := range toRemove {
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "buildpacks:remove", t.App, bp},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
	}

	state.Changed = true
	state.State = StateAbsent
	return state
}

// init registers the BuildpacksTask with the task registry
func init() {
	RegisterTask(&BuildpacksTask{})
}
