package tasks

// OpenrestyPropertyTask manages the openresty configuration for a given dokku application
type OpenrestyPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the openresty configuration should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Property is the name of the openresty property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the openresty property
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the openresty configuration
	State State `required:"true" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// OpenrestyPropertyTaskExample contains an example of an OpenrestyPropertyTask
type OpenrestyPropertyTaskExample struct {
	// Name is the task name holding the OpenrestyPropertyTask description
	Name string `yaml:"-"`

	// OpenrestyPropertyTask is the OpenrestyPropertyTask configuration
	OpenrestyPropertyTask OpenrestyPropertyTask `yaml:"dokku_openresty_property"`
}

// GetName returns the name of the example
func (e OpenrestyPropertyTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the openresty property task
func (t OpenrestyPropertyTask) Doc() string {
	return "Manages the openresty configuration for a given dokku application"
}

// Examples returns the examples for the openresty property task
func (t OpenrestyPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]OpenrestyPropertyTaskExample{
		{
			Name: "Setting the proxy read timeout for an app",
			OpenrestyPropertyTask: OpenrestyPropertyTask{
				App:      "node-js-app",
				Property: "proxy-read-timeout",
				Value:    "120s",
			},
		},
		{
			Name: "Setting the client max body size for an app",
			OpenrestyPropertyTask: OpenrestyPropertyTask{
				App:      "node-js-app",
				Property: "client-max-body-size",
				Value:    "50m",
			},
		},
		{
			Name: "Setting a global openresty property",
			OpenrestyPropertyTask: OpenrestyPropertyTask{
				Global:   true,
				Property: "bind-address-ipv4",
				Value:    "0.0.0.0",
			},
		},
		{
			Name: "Clearing an openresty property",
			OpenrestyPropertyTask: OpenrestyPropertyTask{
				App:      "node-js-app",
				Property: "proxy-read-timeout",
			},
		},
	})
}

// Execute sets or unsets the openresty property
func (t OpenrestyPropertyTask) Execute() TaskOutputState {
	return executeProperty(t.State, t.App, t.Global, t.Property, t.Value, "openresty:set")
}

// init registers the OpenrestyPropertyTask with the task registry
func init() {
	RegisterTask(&OpenrestyPropertyTask{})
}
