package tasks

// NginxPropertyTask manages the nginx configuration for a given dokku application
type NginxPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the nginx configuration should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Property is the name of the nginx property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the nginx property
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the nginx configuration
	State State `required:"true" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// NginxPropertyTaskExample contains an example of a NginxPropertyTask
type NginxPropertyTaskExample struct {
	// Name is the task name holding the NginxPropertyTask description
	Name string `yaml:"-"`

	// NginxPropertyTask is the NginxPropertyTask configuration
	NginxPropertyTask NginxPropertyTask `yaml:"dokku_nginx_property"`
}

// GetName returns the name of the example
func (e NginxPropertyTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the nginx property task
func (t NginxPropertyTask) Doc() string {
	return "Manages the nginx configuration for a given dokku application"
}

// Examples returns the examples for the nginx property task
func (t NginxPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]NginxPropertyTaskExample{
		{
			Name: "Setting the proxy read timeout for an app",
			NginxPropertyTask: NginxPropertyTask{
				App:      "node-js-app",
				Property: "proxy-read-timeout",
				Value:    "120s",
			},
		},
		{
			Name: "Setting the client max body size for an app",
			NginxPropertyTask: NginxPropertyTask{
				App:      "node-js-app",
				Property: "client-max-body-size",
				Value:    "50m",
			},
		},
		{
			Name: "Setting a global nginx property",
			NginxPropertyTask: NginxPropertyTask{
				Global:   true,
				Property: "bind-address-ipv4",
				Value:    "0.0.0.0",
			},
		},
		{
			Name: "Clearing an nginx property",
			NginxPropertyTask: NginxPropertyTask{
				App:      "node-js-app",
				Property: "proxy-read-timeout",
			},
		},
	})
}

// Execute sets or unsets the nginx property
func (t NginxPropertyTask) Execute() TaskOutputState {
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the NginxPropertyTask would produce.
func (t NginxPropertyTask) Plan() PlanResult {
	return planProperty(t.State, t.App, t.Global, t.Property, t.Value, "nginx:set")
}

// init registers the NginxPropertyTask with the task registry
func init() {
	RegisterTask(&NginxPropertyTask{})
}
