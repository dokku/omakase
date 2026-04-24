package tasks

import (
	"errors"
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

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

// DesiredState returns the desired state of the network property
func (t NetworkPropertyTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the network property task
func (t NetworkPropertyTask) Doc() string {
	return "Manages the network property for a given dokku application"
}

// Examples returns the examples for the builder property task
func (t NetworkPropertyTask) Examples() ([]Doc, error) {
	examples := []NetworkPropertyTaskExample{
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
	}

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

// Execute sets or unsets the network property
func (t NetworkPropertyTask) Execute() TaskOutputState {
	if !t.Global && t.App == "" {
		return TaskOutputState{
			Error: errors.New("app is required when global is false"),
		}
	}

	ctx := PropertyContext{
		App:      t.App,
		Global:   t.Global,
		Property: t.Property,
		Value:    t.Value,
	}
	funcMap := map[State]func() TaskOutputState{
		"present": func() TaskOutputState {
			return setProperty("network:set", ctx)
		},
		"absent": func() TaskOutputState {
			return unsetProperty("network:set", ctx)
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

// init registers the NetworkPropertyTask with the task registry
func init() {
	RegisterTask(&NetworkPropertyTask{})
}
