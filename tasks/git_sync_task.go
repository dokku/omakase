package tasks

// GitSyncTask syncs a git repository to a dokku application
type GitSyncTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Repository is the git repository to sync
	Repository string `required:"true" yaml:"repository"`

	// GitRef is the git reference to sync
	GitRef string `required:"false" yaml:"git_ref"`

	// State is the desired state of the git sync
	State State `required:"true" yaml:"state" default:"present"`
}

// DesiredState returns the desired state of the git sync
func (t GitSyncTask) DesiredState() State {
	return t.State
}

// Execute syncs a git repository to a dokku application
func (t GitSyncTask) Execute() TaskOutputState {
	return TaskOutputState{}
}

// init registers the GitSyncTask with the task registry
func init() {
	RegisterTask(&GitSyncTask{})
}
