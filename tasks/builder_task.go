package tasks

// BuilderTask manages the builder configuration for a given dokku application
type BuilderTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the builder configuration should be applied globally
	Global bool `required:"false" yaml:"global"`

	// Property is the name of the builder property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the builder property
	Value string `required:"false" yaml:"value"`

	// State is the desired state of the builder configuration
	State State `required:"true" yaml:"state" default:"present" options:"present,absent"`
}

// DesiredState returns the desired state of the builder configuration
func (t BuilderTask) DesiredState() State {
	return t.State
}

// Execute executes the builder configuration task
func (t BuilderTask) Execute() TaskOutputState {
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

// init registers the BuilderTask with the task registry
func init() {
	RegisterTask(&BuilderTask{})
}
