package tasks

import (
	"omakase/subprocess"

	yaml "gopkg.in/yaml.v3"
)

// GitSyncTask syncs a git repository to a dokku application
type GitSyncTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Repository is the git repository to sync
	Repository string `required:"true" yaml:"repository"`

	// GitRef is the git reference to sync
	GitRef string `required:"false" yaml:"git_ref"`

	// State is the desired state of the git sync
	State State `required:"false" yaml:"state" default:"synced" options:"synced"`
}

// GitSyncTaskExample contains an example of a GitSyncTask
type GitSyncTaskExample struct {
	// Name is the task name holding the GitSyncTask description
	Name string `yaml:"-"`

	// GitSyncTask is the GitSyncTask configuration
	GitSyncTask GitSyncTask `yaml:"git_sync"`
}

// DesiredState returns the desired state of the git sync
func (t GitSyncTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the git sync task
func (t GitSyncTask) Doc() string {
	return "Syncs a git repository to a dokku application"
}

// Examples returns the examples for the builder property task
func (t GitSyncTask) Examples() ([]Doc, error) {
	examples := []GitSyncTaskExample{}

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

// Execute syncs a git repository to a dokku application
func (t GitSyncTask) Execute() TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "unsynced",
	}

	args := []string{
		"git:sync",
	}

	args = append(args, t.App, t.Repository)

	if t.GitRef != "" {
		args = append(args, t.GitRef)
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "synced"
	return state
}

// init registers the GitSyncTask with the task registry
func init() {
	RegisterTask(&GitSyncTask{})
}
