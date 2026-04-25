package tasks

// ChecksToggleTask enables or disables the checks plugin for a given dokku application
type ChecksToggleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the checks plugin should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// State is the desired state of the checks plugin
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// ChecksToggleTaskExample contains an example of a ChecksToggleTask
type ChecksToggleTaskExample struct {
	// Name is the task name holding the ChecksToggleTask description
	Name string `yaml:"-"`

	// ChecksToggleTask is the ChecksToggleTask configuration
	ChecksToggleTask ChecksToggleTask `yaml:"dokku_checks_toggle"`
}

// GetName returns the name of the example
func (e ChecksToggleTaskExample) GetName() string {
	return e.Name
}

// DesiredState returns the desired state of the checks plugin
func (t ChecksToggleTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the checks toggle task
func (t ChecksToggleTask) Doc() string {
	return "Enables or disables the checks plugin for a given dokku application"
}

// Examples returns the examples for the checks toggle task
func (t ChecksToggleTask) Examples() ([]Doc, error) {
	return MarshalExamples([]ChecksToggleTaskExample{
		{
			Name: "Disable the zero downtime deployment",
			ChecksToggleTask: ChecksToggleTask{
				App:   "hello-world",
				State: "absent",
			},
		},
		{
			Name: "Re-enable the zero downtime deployment (enabled by default)",
			ChecksToggleTask: ChecksToggleTask{
				App:   "hello-world",
				State: "present",
			},
		},
	})
}

// Execute enables or disables the checks plugin
func (t ChecksToggleTask) Execute() TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: false,
		App:         t.App,
		Global:      t.Global,
	}
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return enablePlugin("checks:enable", ctx) },
		"absent":  func() TaskOutputState { return disablePlugin("checks:disable", ctx) },
	})
}

// init registers the ChecksToggleTask with the task registry
func init() {
	RegisterTask(&ChecksToggleTask{})
}
