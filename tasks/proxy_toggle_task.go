package tasks

type ProxyToggleTask struct {
	App    string `required:"true" yaml:"app"`
	Global bool   `required:"false" yaml:"global"`
	State  State  `required:"true" yaml:"state" default:"present"`
}

func (t ProxyToggleTask) DesiredState() State {
	return t.State
}

func (t ProxyToggleTask) Execute() TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: false,
		App:         t.App,
		Global:      t.Global,
	}
	funcMap := map[State]func() TaskOutputState{
		"present": func() TaskOutputState {
			return enablePlugin("proxy:enable", ctx)
		},
		"absent": func() TaskOutputState {
			return disablePlugin("proxy:disable", ctx)
		},
	}

	fn := funcMap[t.State]
	return fn()
}

func init() {
	RegisterTask(&ProxyToggleTask{})
}
