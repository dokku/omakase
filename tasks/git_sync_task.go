package tasks

type GitSyncTask struct {
	App        string `required:"true" yaml:"app"`
	Repository string `required:"true" yaml:"repository"`
	GitRef     string `required:"false" yaml:"git_ref"`
	State      State  `required:"true" yaml:"state" default:"present"`
}

func (t GitSyncTask) DesiredState() State {
	return t.State
}

func (t GitSyncTask) Execute() TaskOutputState {
	return TaskOutputState{}
}

func init() {
	RegisterTask(&GitSyncTask{})
}
