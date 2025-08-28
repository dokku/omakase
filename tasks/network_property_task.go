package tasks

// NetworkPropertyTask manages the network property for a given dokku application
type NetworkPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the network property should be applied globally
	Global bool `required:"false" yaml:"global"`

	// Property is the name of the network property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value of the network property to set
	Value string `required:"false" yaml:"value"`

	// State is the desired state of the network property
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// DesiredState returns the desired state of the network property
func (t NetworkPropertyTask) DesiredState() State {
	return t.State
}

// Execute sets or unsets the network property
func (t NetworkPropertyTask) Execute() TaskOutputState {
	ctx := PropertyContext{
		App:      t.App,
		Global:   t.Global,
		Property: t.Property,
		Value:    t.Value,
	}
	funcMap := map[State]func() TaskOutputState{
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

// init registers the NetworkPropertyTask with the task registry
func init() {
	RegisterTask(&NetworkPropertyTask{})
}
