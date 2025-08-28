package tasks

type BuilderPropertyTask struct {
	App      string `required:"true" yaml:"app"`
	Global   bool   `required:"false" yaml:"global"`
	Property string `required:"true" yaml:"property"`
	Value    string `required:"false" yaml:"value"`
	State    State  `required:"true" yaml:"state" default:"present"`
}

func (t BuilderPropertyTask) DesiredState() State {
	return t.State
}

func (t BuilderPropertyTask) Execute() TaskOutputState {
	ctx := PropertyContext{
		App:      t.App,
		Global:   t.Global,
		Property: t.Property,
		Value:    t.Value,
	}
	funcMap := map[State]func() TaskOutputState{
		"present": func() TaskOutputState {
			return setProperty("builder:set", ctx)
		},
		"absent": func() TaskOutputState {
			return unsetProperty("builder:set", ctx)
		},
	}

	fn := funcMap[t.State]
	return fn()
}

func init() {
	RegisterTask(&BuilderPropertyTask{})
}
