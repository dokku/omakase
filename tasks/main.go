package tasks

import (
	"crypto/rand"
	"fmt"
	"io"
	"reflect"
	"strings"

	sigil "github.com/gliderlabs/sigil"
	"github.com/gobuffalo/flect"
	defaults "github.com/mcuadros/go-defaults"
	yaml "gopkg.in/yaml.v3"
)

// State represents the desired state of a task
type State string

// State constants
const (
	// StatePresent represents the present state
	StatePresent State = "present"
	// StateAbsent represents the absent state
	StateAbsent State = "absent"
	// StateDeployed represents the deployed state
	StateDeployed State = "deployed"
	// StateSet represents the set state
	StateSet State = "set"
	// StateClear represents the clear state
	StateClear State = "clear"
)

// Recipe represents a recipe for a task
type Recipe []struct {
	// Inputs are the inputs for the task
	Inputs []Input `yaml:"inputs,omitempty"`

	// Tasks are the tasks for the recipe
	Tasks []map[string]interface{} `yaml:"tasks,omitempty"`
}

// Input represents an input for a task
type Input struct {
	// Name is the name of the input
	Name string `yaml:"name"`

	// Default is the default value of the input
	Default string `yaml:"default"`

	// Description is the description of the input
	Description string `yaml:"description"`

	// Required is a flag indicating if the input is required
	Required bool `yaml:"required"`

	// Type is the type of the input
	Type string `yaml:"type"`

	// value is the value of the input
	value string
}

// TaskOutputState represents the output of a task
type TaskOutputState struct {
	// Changed is a flag indicating if the task was changed
	Changed bool

	// DesiredState is the desired state of the task
	DesiredState State

	// Error is the error of the task
	Error error

	// Message is the message of the task
	Message string

	// Meta is the meta of the task
	Meta struct{}

	// State is the state of the task
	State State
}

// Task represents a task
type Task interface {
	// Doc returns the docblock for the task
	Doc() string

	// Examples returns the examples for the task
	Examples() ([]Doc, error)

	// Execute executes the task
	Execute() TaskOutputState
}

// Global registry for Tasks.
var RegisteredTasks map[string]Task

// RegisterTask registers a task
func RegisterTask(t Task) {
	if len(RegisteredTasks) == 0 {
		RegisteredTasks = make(map[string]Task)
	}

	var name string
	if t := reflect.TypeOf(t); t.Kind() == reflect.Ptr {
		name = "*" + t.Elem().Name()
	} else {
		name = t.Name()
	}

	name = flect.Underscore(name)
	RegisteredTasks[fmt.Sprintf("dokku_%s", strings.TrimSuffix(name, "_task"))] = t
}

// SetValue sets the value of the input
func (i *Input) SetValue(value string) error {
	i.value = value
	return nil
}

// HasValue returns true if the input has a value
func (i Input) HasValue() bool {
	return i.value != ""
}

// GetValue returns the value of the input
func (i Input) GetValue() string {
	return i.value
}

// GetTasks gets the tasks from the data
// todo: use a slice instead of a map
func GetTasks(data []byte, context map[string]interface{}) (OrderedStringTaskMap, error) {
	tasks := OrderedStringTaskMap{}
	render, err := sigil.Execute(data, context, "tasks")
	if err != nil {
		return tasks, fmt.Errorf("re-render error: %v", err.Error())
	}

	out, err := io.ReadAll(&render)
	if err != nil {
		return tasks, fmt.Errorf("read error: %v", err.Error())
	}

	recipe := Recipe{}
	if err := yaml.Unmarshal([]byte(out), &recipe); err != nil {
		return tasks, fmt.Errorf("unmarshal error: %v", err.Error())
	}

	i := 0
	validTasks := make([]string, len(RegisteredTasks))
	for k := range RegisteredTasks {
		validTasks[i] = k
		i++
	}

	if len(recipe) == 0 {
		return tasks, fmt.Errorf("parse error: no recipe found in tasks file")
	}

	for i, t := range recipe[0].Tasks {
		if len(t) > 2 {
			return tasks, fmt.Errorf("task parse error: task #%d has too many properties - expected=2 actual=%d", i+1, len(t))
		}

		name, ok := t["name"]
		if len(t) == 2 && !ok {
			keys := make([]string, len(t))

			j := 0
			for k := range t {
				keys[j] = k
				j++
			}

			return tasks, fmt.Errorf("task parse error: task #%d has an unexpected property - properties=%v", i+1, keys)
		}

		if !ok {
			b := make([]byte, 8)
			if _, err := rand.Read(b); err != nil {
				return tasks, fmt.Errorf("task parse error: task #%d had no task name and there was a failure to generate random task name - %s", i+1, err)
			}
			name = fmt.Sprintf("task #%d %X", i+1, b)
		}

		detected := false
		for taskName, registeredTask := range RegisteredTasks {
			config, ok := t[taskName]
			if !ok {
				continue
			}

			marshaled, err := yaml.Marshal(config)
			if err != nil {
				return tasks, fmt.Errorf("task parse error: task #%d failed to marshal config to yaml - %s", i+1, err)
			}

			v := reflect.New(reflect.TypeOf(registeredTask))
			if err := yaml.Unmarshal([]byte(marshaled), v.Interface()); err != nil {
				return tasks, fmt.Errorf("task parse error: task #%d failed to decode to %s - %s", i+1, taskName, err)
			}

			task := v.Elem().Interface().(Task)
			defaults.SetDefaults(task)
			tasks.Set(name.(string), task)
			detected = true
			break
		}

		if !detected {
			return tasks, fmt.Errorf("task parse error: task #%d was not a valid task - valid_tasks=%v", i+1, validTasks)
		}
	}

	return tasks, nil
}
