package tasks

// TraefikPropertyTask manages the traefik configuration for a given dokku application
type TraefikPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the traefik configuration should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Property is the name of the traefik property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the traefik property
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the traefik configuration
	State State `required:"true" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// TraefikPropertyTaskExample contains an example of a TraefikPropertyTask
type TraefikPropertyTaskExample struct {
	// Name is the task name holding the TraefikPropertyTask description
	Name string `yaml:"-"`

	// TraefikPropertyTask is the TraefikPropertyTask configuration
	TraefikPropertyTask TraefikPropertyTask `yaml:"dokku_traefik_property"`
}

// GetName returns the name of the example
func (e TraefikPropertyTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the traefik property task
func (t TraefikPropertyTask) Doc() string {
	return "Manages the traefik configuration for a given dokku application"
}

// Examples returns the examples for the traefik property task
func (t TraefikPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]TraefikPropertyTaskExample{
		{
			Name: "Setting the letsencrypt email for an app",
			TraefikPropertyTask: TraefikPropertyTask{
				App:      "node-js-app",
				Property: "letsencrypt-email",
				Value:    "admin@example.com",
			},
		},
		{
			Name: "Setting the log level globally",
			TraefikPropertyTask: TraefikPropertyTask{
				Global:   true,
				Property: "log-level",
				Value:    "INFO",
			},
		},
		{
			Name: "Clearing the letsencrypt email for an app",
			TraefikPropertyTask: TraefikPropertyTask{
				App:      "node-js-app",
				Property: "letsencrypt-email",
			},
		},
	})
}

// Execute sets or unsets the traefik property
func (t TraefikPropertyTask) Execute() TaskOutputState {
	return executeProperty(t.State, t.App, t.Global, t.Property, t.Value, "traefik:set")
}

// Plan reports the drift the TraefikPropertyTask would produce.
func (t TraefikPropertyTask) Plan() PlanResult {
	return planProperty(t.State, t.App, t.Global, t.Property, t.Value, "traefik:set")
}

// init registers the TraefikPropertyTask with the task registry
func init() {
	RegisterTask(&TraefikPropertyTask{})
}
