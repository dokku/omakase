package tasks

// DomainsToggleTask enables or disables the domains plugin for a given dokku application
type DomainsToggleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the domains plugin should be applied globally
	Global bool `required:"false" yaml:"global"`

	// State is the desired state of the domains plugin
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// DomainsToggleTaskExample contains an example of a DomainsToggleTask
type DomainsToggleTaskExample struct {
	// Name is the task name holding the DomainsToggleTask description
	Name string `yaml:"-"`

	// DomainsToggleTask is the DomainsToggleTask configuration
	DomainsToggleTask DomainsToggleTask `yaml:"dokku_domains_toggle"`
}

// GetName returns the name of the example
func (e DomainsToggleTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the domains toggle task
func (t DomainsToggleTask) Doc() string {
	return "Enables or disables the domains plugin for a given dokku application"
}

// Examples returns the examples for the domains toggle task
func (t DomainsToggleTask) Examples() ([]Doc, error) {
	return MarshalExamples([]DomainsToggleTaskExample{})
}

// Execute enables or disables the domains plugin
func (t DomainsToggleTask) Execute() TaskOutputState {
	return executeToggle(t.State, t.App, t.Global, false, "domains:enable", "domains:disable")
}

// Plan reports the drift the DomainsToggleTask would produce.
func (t DomainsToggleTask) Plan() PlanResult {
	return planToggle(t.State, t.App, t.Global, false, "domains:enable", "domains:disable")
}

// init registers the DomainsToggleTask with the task registry
func init() {
	RegisterTask(&DomainsToggleTask{})
}
