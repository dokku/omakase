package tasks

import (
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

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
	DomainsToggleTask DomainsToggleTask `yaml:"domains_toggle"`
}

// DesiredState returns the desired state of the domains plugin
func (t DomainsToggleTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the domains toggle task
func (t DomainsToggleTask) Doc() string {
	return "Enables or disables the domains plugin for a given dokku application"
}

// Examples returns the examples for the builder property task
func (t DomainsToggleTask) Examples() ([]Doc, error) {
	examples := []DomainsToggleTaskExample{}

	var output []Doc
	for _, example := range examples {
		b, err := yaml.Marshal(example)
		if err != nil {
			return nil, err
		}

		output = append(output, Doc{
			Name:      example.Name,
			Codeblock: string(b),
		})
	}

	return output, nil
}

// Execute enables or disables the domains plugin
func (t DomainsToggleTask) Execute() TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: false,
		App:         t.App,
		Global:      t.Global,
	}
	funcMap := map[State]func() TaskOutputState{
		"present": func() TaskOutputState {
			return enablePlugin("domains:enable", ctx)
		},
		"absent": func() TaskOutputState {
			return disablePlugin("domains:disable", ctx)
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

// init registers the DomainsToggleTask with the task registry
func init() {
	RegisterTask(&DomainsToggleTask{})
}
