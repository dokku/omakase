package tasks

type NetworkPropertyTask struct {
	App      string `required:"true" yaml:"app"`
	Global   bool   `required:"false" yaml:"global"`
	Property string `required:"true" yaml:"property"`
	Value    string `required:"false" yaml:"value"`
	State    string `required:"true" yaml:"state" default:"present"`
}

func (t NetworkPropertyTask) DesiredState() string {
	return t.State
}

func (t NetworkPropertyTask) Execute() TaskOutputState {
	ctx := PropertyContext{
		App:      t.App,
		Global:   t.Global,
		Property: t.Property,
		Value:    t.Value,
	}
	funcMap := map[string]func() TaskOutputState{
		"present": func() TaskOutputState {
			return setProperty("network:set", ctx)
		},
		"absent": func() TaskOutputState {
			return unsetProperty("network:set", ctx)
		},
	}

	fn := funcMap[t.State]
	return fn()
}

func init() {
	RegisterTask(&NetworkPropertyTask{})
}
