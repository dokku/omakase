package main

type AppTask struct {
	App   string `required:"true" yaml:"app"`
	State string `required:"true" yaml:"state" default:"present"`
}

func (t AppTask) DesiredState() string {
	return t.State
}

func (t AppTask) Execute() (string, error) {
	return "", nil
}

func (t AppTask) NeedsExecution() bool {
	return true
}

func (t *AppTask) SetDefaultDesiredState(state string) {
	if t.State == "" {
		t.State = state
	}
}
