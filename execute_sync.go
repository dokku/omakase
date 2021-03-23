package main

type SyncTask struct {
	App        string `required:"true" yaml:"app"`
	Repository string `required:"true" yaml:"repository"`
	State      string `required:"true" yaml:"state" default:"present"`
}

func (t SyncTask) DesiredState() string {
	return t.State
}

func (t SyncTask) NeedsExecution() bool {
	return true
}

func (t SyncTask) Execute() (string, error) {
	return "", nil
}
