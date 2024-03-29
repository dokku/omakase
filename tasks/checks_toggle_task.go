package tasks

type ChecksToggleTask struct {
	App    string `required:"true" yaml:"app"`
	Global bool   `required:"false" yaml:"global"`
	State  string `required:"true" yaml:"state" default:"present"`
}

func (t ChecksToggleTask) DesiredState() string {
	return t.State
}

func (t ChecksToggleTask) Execute() TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: false,
		App:         t.App,
		Global:      t.Global,
	}
	funcMap := map[string]func() TaskOutputState{
		"present": func() TaskOutputState {
			return enablePlugin("checks:enable", ctx)
		},
		"absent": func() TaskOutputState {
			return disablePlugin("checks:disable", ctx)
		},
	}

	fn := funcMap[t.State]
	return fn()
}

func init() {
	RegisterTask(&ChecksToggleTask{})
}
