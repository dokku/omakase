package tasks

// ProxyToggleTask manages the proxy for a given dokku application
type ProxyToggleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the proxy should be applied globally
	Global bool `required:"false" yaml:"global"`

	// State is the desired state of the proxy
	State State `required:"true" yaml:"state" default:"present"`
}

// DesiredState returns the desired state of the proxy
func (t ProxyToggleTask) DesiredState() State {
	return t.State
}

// Execute enables or disables the proxy
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

// init registers the ProxyToggleTask with the task registry
func init() {
	RegisterTask(&ProxyToggleTask{})
}
