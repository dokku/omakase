package tasks

type BuilderTask struct {
	App      string `required:"true" yaml:"app"`
	Global   bool   `required:"false" yaml:"global"`
	Property string `required:"true" yaml:"property"`
	Value    string `required:"false" yaml:"value"`
	State    string `required:"true" yaml:"state" default:"present"`
}

func (t BuilderTask) DesiredState() string {
	return t.State
}

func (t BuilderTask) Execute() TaskOutputState {
	ctx := PropertyContext{
		App:      t.App,
		Global:   t.Global,
		Property: t.Property,
		Value:    t.Value,
	}
	funcMap := map[string]func() TaskOutputState{
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
	RegisterTask(&BuilderTask{})
}
