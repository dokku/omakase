package tasks

// NetworkPropertyTask manages the network property for a given dokku application
type NetworkPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the network property should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Property is the name of the network property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value of the network property to set
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the network property
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// NetworkPropertyTaskExample contains an example of a NetworkPropertyTask
type NetworkPropertyTaskExample struct {
	// Name is the task name holding the NetworkPropertyTask description
	Name string `yaml:"-"`

	// NetworkPropertyTask is the NetworkPropertyTask configuration
	NetworkPropertyTask NetworkPropertyTask `yaml:"dokku_network_property"`
}

// GetName returns the name of the example
func (e NetworkPropertyTaskExample) GetName() string {
	return e.Name
}

// DesiredState returns the desired state of the network property
func (t NetworkPropertyTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the network property task
func (t NetworkPropertyTask) Doc() string {
	return "Manages the network property for a given dokku application"
}

// Examples returns the examples for the network property task
func (t NetworkPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]NetworkPropertyTaskExample{
		{
			Name: "Associates a network after a container is created but before it is started",
			NetworkPropertyTask: NetworkPropertyTask{
				App:      "hello-world",
				Property: "attach-post-create",
				Value:    "example-network",
			},
		},
		{
			Name: "Associates the network at container creation",
			NetworkPropertyTask: NetworkPropertyTask{
				App:      "hello-world",
				Property: "initial-network",
				Value:    "example-network",
			},
		},
		{
			Name: "Setting a global network property",
			NetworkPropertyTask: NetworkPropertyTask{
				Global:   true,
				Property: "attach-post-create",
				Value:    "example-network",
			},
		},
		{
			Name: "Clearing a network property",
			NetworkPropertyTask: NetworkPropertyTask{
				App:      "hello-world",
				Property: "attach-post-create",
			},
		},
	})
}

// Execute sets or unsets the network property
func (t NetworkPropertyTask) Execute() TaskOutputState {
	return executeProperty(t.State, t.App, t.Global, t.Property, t.Value, "network:set")
}

// init registers the NetworkPropertyTask with the task registry
func init() {
	RegisterTask(&NetworkPropertyTask{})
}
