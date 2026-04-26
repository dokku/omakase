package tasks

// ProxyToggleTask manages the proxy for a given dokku application
type ProxyToggleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the proxy should be applied globally
	Global bool `required:"false" yaml:"global"`

	// State is the desired state of the proxy
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// ProxyToggleTaskExample contains an example of a ProxyToggleTask
type ProxyToggleTaskExample struct {
	// Name is the task name holding the ProxyToggleTask description
	Name string `yaml:"-"`

	// ProxyToggleTask is the ProxyToggleTask configuration
	ProxyToggleTask ProxyToggleTask `yaml:"dokku_proxy_toggle"`
}

// GetName returns the name of the example
func (e ProxyToggleTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the proxy toggle task
func (t ProxyToggleTask) Doc() string {
	return "Enables or disables the proxy plugin for a given dokku application"
}

// Examples returns the examples for the proxy toggle task
func (t ProxyToggleTask) Examples() ([]Doc, error) {
	return MarshalExamples([]ProxyToggleTaskExample{})
}

// Execute enables or disables the proxy
func (t ProxyToggleTask) Execute() TaskOutputState {
	return executeToggle(t.State, t.App, t.Global, false, "proxy:enable", "proxy:disable")
}

// init registers the ProxyToggleTask with the task registry
func init() {
	RegisterTask(&ProxyToggleTask{})
}
