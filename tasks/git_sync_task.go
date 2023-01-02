package tasks

type GitSyncTask struct {
	App        string `required:"true" yaml:"app"`
	Repository string `required:"true" yaml:"repository"`
	GitRef     string `required:"false" yaml:"repository"`
	State      string `required:"true" yaml:"state" default:"present"`
}

func (t GitSyncTask) DesiredState() string {
	return t.State
}

func (t GitSyncTask) Execute() (string, error) {
	return "", nil
}

func init() {
	RegisterTask(&GitSyncTask{})
}
