package tasks

// ChecksToggleTask enables or disables the checks plugin for a given dokku application
type ChecksToggleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the checks plugin should be applied globally
	Global bool `required:"false" yaml:"global"`

	// State is the desired state of the checks plugin
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// DesiredState returns the desired state of the checks plugin
func (t ChecksToggleTask) DesiredState() State {
	return t.State
}

// Execute enables or disables the checks plugin
func (t ChecksToggleTask) Execute() TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: false,
		App:         t.App,
		Global:      t.Global,
	}
	funcMap := map[State]func() TaskOutputState{
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

// init registers the ChecksToggleTask with the task registry
func init() {
	RegisterTask(&ChecksToggleTask{})
}
