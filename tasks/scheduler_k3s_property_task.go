package tasks

// SchedulerK3sPropertyTask manages the scheduler-k3s configuration for a given dokku application
type SchedulerK3sPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the scheduler-k3s configuration should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Property is the name of the scheduler-k3s property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the scheduler-k3s property
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the scheduler-k3s configuration
	State State `required:"true" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// SchedulerK3sPropertyTaskExample contains an example of a SchedulerK3sPropertyTask
type SchedulerK3sPropertyTaskExample struct {
	// Name is the task name holding the SchedulerK3sPropertyTask description
	Name string `yaml:"-"`

	// SchedulerK3sPropertyTask is the SchedulerK3sPropertyTask configuration
	SchedulerK3sPropertyTask SchedulerK3sPropertyTask `yaml:"dokku_scheduler_k3s_property"`
}

// GetName returns the name of the example
func (e SchedulerK3sPropertyTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the scheduler-k3s property task
func (t SchedulerK3sPropertyTask) Doc() string {
	return "Manages the scheduler-k3s configuration for a given dokku application"
}

// Examples returns the examples for the scheduler-k3s property task
func (t SchedulerK3sPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]SchedulerK3sPropertyTaskExample{
		{
			Name: "Setting the deploy timeout for an app",
			SchedulerK3sPropertyTask: SchedulerK3sPropertyTask{
				App:      "node-js-app",
				Property: "deploy-timeout",
				Value:    "300s",
			},
		},
		{
			Name: "Setting the namespace for an app",
			SchedulerK3sPropertyTask: SchedulerK3sPropertyTask{
				App:      "node-js-app",
				Property: "namespace",
				Value:    "production",
			},
		},
		{
			Name: "Setting the letsencrypt prod email globally",
			SchedulerK3sPropertyTask: SchedulerK3sPropertyTask{
				Global:   true,
				Property: "letsencrypt-email-prod",
				Value:    "admin@example.com",
			},
		},
		{
			Name: "Clearing the namespace for an app",
			SchedulerK3sPropertyTask: SchedulerK3sPropertyTask{
				App:      "node-js-app",
				Property: "namespace",
			},
		},
	})
}

// Execute sets or unsets the scheduler-k3s property
func (t SchedulerK3sPropertyTask) Execute() TaskOutputState {
	return executeProperty(t.State, t.App, t.Global, t.Property, t.Value, "scheduler-k3s:set")
}

// init registers the SchedulerK3sPropertyTask with the task registry
func init() {
	RegisterTask(&SchedulerK3sPropertyTask{})
}
