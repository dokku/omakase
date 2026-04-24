package tasks

import (
	"errors"
	"fmt"
	"omakase/subprocess"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ResourceLimitTask manages the resource limits for a given dokku application
type ResourceLimitTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// ProcessType is the process type to set resource limits for
	ProcessType string `required:"false" yaml:"process_type,omitempty"`

	// Resources is a map of resource type to quantity
	Resources map[string]string `yaml:"resources"`

	// ClearBefore clears all resource limits before applying new ones
	ClearBefore bool `yaml:"clear_before" default:"false"`

	// State is the desired state of the resource limits
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// ResourceLimitTaskExample contains an example of a ResourceLimitTask
type ResourceLimitTaskExample struct {
	// Name is the task name holding the ResourceLimitTask description
	Name string `yaml:"-"`

	// ResourceLimitTask is the ResourceLimitTask configuration
	ResourceLimitTask ResourceLimitTask `yaml:"resource_limit"`
}

// DesiredState returns the desired state of the resource limits
func (t ResourceLimitTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the resource limit task
func (t ResourceLimitTask) Doc() string {
	return "Manages the resource limits for a given dokku application"
}

// Examples returns the examples for the resource limit task
func (t ResourceLimitTask) Examples() ([]Doc, error) {
	examples := []ResourceLimitTaskExample{
		{
			Name: "Set CPU and memory limits",
			ResourceLimitTask: ResourceLimitTask{
				App: "hello-world",
				Resources: map[string]string{
					"cpu":    "100",
					"memory": "256",
				},
			},
		},
		{
			Name: "Set memory limit for web process type",
			ResourceLimitTask: ResourceLimitTask{
				App:         "hello-world",
				ProcessType: "web",
				Resources: map[string]string{
					"memory": "512",
				},
			},
		},
		{
			Name: "Clear all resource limits",
			ResourceLimitTask: ResourceLimitTask{
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

// Execute sets or clears the resource limits for a given dokku application
func (t ResourceLimitTask) Execute() TaskOutputState {
	funcMap := map[State]func(ResourceLimitTask) TaskOutputState{
		"present": setResourceLimit,
		"absent":  clearResourceLimit,
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
	return fn(t)
}

// getResourceLimits retrieves the current resource limits for a given dokku application
func getResourceLimits(t ResourceLimitTask) (map[string]string, error) {
	args := []string{
		"--quiet",
		"resource:limit",
	}

	if t.ProcessType != "" {
		args = append(args, "--process-type", t.ProcessType)
	}

	args = append(args, t.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return nil, err
	}

	limits := map[string]string{}
	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			limits[key] = value
		}
	}

	return limits, nil
}

// setResourceLimit sets the resource limits for a given dokku application
func setResourceLimit(t ResourceLimitTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	if t.ClearBefore {
		err := execResourceLimitClear(t)
		if err != nil {
			state.Error = err
			state.Message = err.Error()
			return state
		}
		state.Changed = true
	}

	if !t.ClearBefore {
		currentLimits, err := getResourceLimits(t)
		if err != nil {
			state.Error = err
			state.Message = err.Error()
			return state
		}

		// validate that all requested resources exist
		for k := range t.Resources {
			if _, ok := currentLimits[k]; !ok {
				state.Error = fmt.Errorf("unknown resource %s, valid resources: %v", k, mapKeys(currentLimits))
				state.Message = state.Error.Error()
				return state
			}
		}

		hasChanged := false
		for k, v := range t.Resources {
			if currentLimits[k] != v {
				hasChanged = true
				break
			}
		}

		if !hasChanged {
			state.State = "present"
			return state
		}
	}

	args := []string{
		"resource:limit",
	}

	for key, value := range t.Resources {
		args = append(args, fmt.Sprintf("--%s", key), value)
	}

	if t.ProcessType != "" {
		args = append(args, "--process-type", t.ProcessType)
	}

	args = append(args, t.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "present"
	return state
}

// clearResourceLimit clears the resource limits for a given dokku application
func clearResourceLimit(t ResourceLimitTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	currentLimits, err := getResourceLimits(t)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	hasLimits := false
	for _, v := range currentLimits {
		if v != "" && v != "0" {
			hasLimits = true
			break
		}
	}

	if !hasLimits {
		state.State = "absent"
		return state
	}

	err = execResourceLimitClear(t)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// execResourceLimitClear executes the dokku resource:limit-clear command
func execResourceLimitClear(t ResourceLimitTask) error {
	args := []string{
		"resource:limit-clear",
	}

	if t.ProcessType != "" {
		args = append(args, "--process-type", t.ProcessType)
	}

	args = append(args, t.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return fmt.Errorf("%s", result.StderrContents())
	}

	return nil
}

// mapKeys returns the keys of a map as a sorted slice
func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// init registers the ResourceLimitTask with the task registry
func init() {
	RegisterTask(&ResourceLimitTask{})
}
