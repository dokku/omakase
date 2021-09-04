package main

type SyncTask struct {
	App        string `required:"true" yaml:"app"`
	Repository string `required:"true" yaml:"repository"`
	State      string `required:"true" yaml:"state" default:"present"`
}

func (t SyncTask) DesiredState() string {
	return t.State
}

func (t SyncTask) Execute() (string, error) {
	return "", nil
}

func (t SyncTask) NeedsExecution() bool {
	return true
}

func (t *SyncTask) SetDefaultDesiredState(state string) {
	if t.State == "" {
		t.State = state
	}
}
