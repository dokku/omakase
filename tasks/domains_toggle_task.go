package tasks

// DomainsToggleTask enables or disables the domains plugin for a given dokku application
type DomainsToggleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the domains plugin should be applied globally
	Global bool `required:"false" yaml:"global"`

	// State is the desired state of the domains plugin
	State State `required:"true" yaml:"state" default:"present"`
}

// DesiredState returns the desired state of the domains plugin
func (t DomainsToggleTask) DesiredState() State {
	return t.State
}

// Execute enables or disables the domains plugin
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

// init registers the DomainsToggleTask with the task registry
func init() {
	RegisterTask(&DomainsToggleTask{})
}
