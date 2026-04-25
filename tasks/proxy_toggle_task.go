package tasks

import (
	"fmt"
)

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

// DesiredState returns the desired state of the proxy
func (t ProxyToggleTask) DesiredState() State {
	return t.State
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

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn()
}

// init registers the ProxyToggleTask with the task registry
func init() {
	RegisterTask(&ProxyToggleTask{})
}
