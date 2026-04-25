package tasks

import (
	"errors"
)

// BuilderTask manages the builder configuration for a given dokku application
type BuilderPropertyTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the builder configuration should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Property is the name of the builder property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the builder property
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the builder configuration
	State State `required:"true" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// BuilderPropertyTaskExample contains an example of a BuilderPropertyTask
type BuilderPropertyTaskExample struct {
	// Name is the task name holding the BuilderPropertyTask description
	Name string `yaml:"-"`

	// BuilderPropertyTask is the BuilderPropertyTask configuration
	BuilderPropertyTask BuilderPropertyTask `yaml:"dokku_builder_property"`
}

// GetName returns the name of the example
func (e BuilderPropertyTaskExample) GetName() string {
	return e.Name
}

// DesiredState returns the desired state of the builder configuration
func (t BuilderPropertyTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the builder property task
func (t BuilderPropertyTask) Doc() string {
	return "Manages the builder configuration for a given dokku application"
}

// Examples returns the examples for the builder property task
func (t BuilderPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]BuilderPropertyTaskExample{
		{
			Name: "Overriding the auto-selected builder",
			BuilderPropertyTask: BuilderPropertyTask{
				App:      "node-js-app",
				Property: "selected",
				Value:    "dockerfile",
			},
		},
		{
			Name: "Setting the builder to the default value",
			BuilderPropertyTask: BuilderPropertyTask{
				App:      "node-js-app",
				Property: "selected",
			},
		},
		{
			Name: "Changing the build build directory",
			BuilderPropertyTask: BuilderPropertyTask{
				App:      "monorepo",
				Property: "build-dir",
				Value:    "backend",
			},
		},
		{
			Name: "Overriding the auto-selected builder globally",
			BuilderPropertyTask: BuilderPropertyTask{
				Global:   true,
				Property: "selected",
				Value:    "herokuish",
			},
		},
	})
}

// Execute executes the builder configuration task
func (t BuilderPropertyTask) Execute() TaskOutputState {
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
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return setProperty("builder:set", ctx) },
		"absent":  func() TaskOutputState { return unsetProperty("builder:set", ctx) },
	})
}

// init registers the BuilderTask with the task registry
func init() {
	RegisterTask(&BuilderPropertyTask{})
}
