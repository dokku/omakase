package tasks

import (
	"errors"
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

// ResourceReserveTask manages the resource reservations for a given dokku application
type ResourceReserveTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// ProcessType is the process type to set resource reservations for
	ProcessType string `required:"false" yaml:"process_type,omitempty"`

	// Resources is a map of resource type to quantity
	Resources map[string]string `yaml:"resources"`

	// ClearBefore clears all resource reservations before applying new ones
	ClearBefore bool `yaml:"clear_before" default:"false"`

	// State is the desired state of the resource reservations
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// ResourceReserveTaskExample contains an example of a ResourceReserveTask
type ResourceReserveTaskExample struct {
	// Name is the task name holding the ResourceReserveTask description
	Name string `yaml:"-"`

	// ResourceReserveTask is the ResourceReserveTask configuration
	ResourceReserveTask ResourceReserveTask `yaml:"dokku_resource_reserve"`
}

// DesiredState returns the desired state of the resource reservations
func (t ResourceReserveTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the resource reserve task
func (t ResourceReserveTask) Doc() string {
	return "Manages the resource reservations for a given dokku application"
}

// Examples returns the examples for the resource reserve task
func (t ResourceReserveTask) Examples() ([]Doc, error) {
	examples := []ResourceReserveTaskExample{
		{
			Name: "Set CPU and memory reservations",
			ResourceReserveTask: ResourceReserveTask{
				App: "hello-world",
				Resources: map[string]string{
					"cpu":    "100",
					"memory": "256",
				},
			},
		},
		{
			Name: "Set memory reservation for web process type",
			ResourceReserveTask: ResourceReserveTask{
				App:         "hello-world",
				ProcessType: "web",
				Resources: map[string]string{
					"memory": "512",
				},
			},
		},
		{
			Name: "Clear all resource reservations",
			ResourceReserveTask: ResourceReserveTask{
				App:   "hello-world",
				State: StateAbsent,
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

// Execute sets or clears the resource reservations for a given dokku application
func (t ResourceReserveTask) Execute() TaskOutputState {
	funcMap := map[State]func(string, ResourceContext) TaskOutputState{
		"present": setResource,
		"absent":  clearResource,
	}

	if t.State == StatePresent && len(t.Resources) == 0 {
		return TaskOutputState{
			Error:   errors.New("resources are required when state is present"),
			Message: "resources are required when state is present",
		}
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}

	rctx := ResourceContext{
		App:         t.App,
		ProcessType: t.ProcessType,
		Resources:   t.Resources,
		ClearBefore: t.ClearBefore,
	}

	return fn("resource:reserve", rctx)
}

// init registers the ResourceReserveTask with the task registry
func init() {
	RegisterTask(&ResourceReserveTask{})
}
