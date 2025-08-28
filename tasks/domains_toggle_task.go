package tasks

type DomainsToggleTask struct {
	App    string `required:"true" yaml:"app"`
	Global bool   `required:"false" yaml:"global"`
	State  State  `required:"true" yaml:"state" default:"present"`
}

func (t DomainsToggleTask) DesiredState() State {
	return t.State
}

func (t DomainsToggleTask) Execute() TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: false,
		App:         t.App,
		Global:      t.Global,
	}
	funcMap := map[State]func() TaskOutputState{
		"present": func() TaskOutputState {
			return enablePlugin("domains:enable", ctx)
		},
		"absent": func() TaskOutputState {
			return disablePlugin("domains:disable", ctx)
		},
	}

	fn := funcMap[t.State]
	return fn()
}

func init() {
	RegisterTask(&DomainsToggleTask{})
}
