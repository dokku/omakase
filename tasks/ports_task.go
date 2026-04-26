package tasks

import (
	"errors"
	"fmt"
	"github.com/dokku/docket/subprocess"
	"strconv"
	"strings"
)

// PortsTask manages the ports for a given dokku application
type PortsTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// PortMappings are the port mappings to set
	PortMappings []PortMapping `required:"true" yaml:"port_mappings"`

	// State is the desired state of the ports
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// PortsTaskExample contains an example of a PortsTask
type PortsTaskExample struct {
	// Name is the task name holding the PortsTask description
	Name string `yaml:"-"`

	// PortsTask is the PortsTask configuration
	PortsTask PortsTask `yaml:"dokku_ports"`
}

// GetName returns the name of the example
func (e PortsTaskExample) GetName() string {
	return e.Name
}

// PortMapping represents a port mapping
type PortMapping struct {
	// Scheme is the scheme of the port mapping
	Scheme string `required:"true" yaml:"scheme"`

	// Host is the host of the port mapping
	Host int `required:"true" yaml:"host"`

	// Container is the container of the port mapping
	Container int `required:"true" yaml:"container"`
}

// String returns the string representation of the port mapping
func (p PortMapping) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Scheme, p.Host, p.Container)
}

// Doc returns the docblock for the ports task
func (t PortsTask) Doc() string {
	return "Manages the ports for a given dokku application"
}

// Examples returns the examples for the ports task
func (t PortsTask) Examples() ([]Doc, error) {
	return MarshalExamples([]PortsTaskExample{})
}

// Execute sets or unsets the ports
func (t PortsTask) Execute() TaskOutputState {
	// todo: add port mapping validation
	if len(t.PortMappings) == 0 {
		return TaskOutputState{
			Changed: false,
			State:   "absent",
			Error:   errors.New("no port mappings provided"),
			Message: "no port mappings provided",
		}
	}

	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return setPorts(t.App, t.PortMappings) },
		"absent":  func() TaskOutputState { return unsetPorts(t.App, t.PortMappings) },
	})
}

// getPorts gets the ports for a given app
func getPorts(appName string) map[string]PortMapping {
	// todo: update dokku to add proper json output for this
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"ports:report",
			appName,
			"--ports-map",
		},
	})
	if err != nil {
		return map[string]PortMapping{}
	}

	portMappings := map[string]PortMapping{}
	for _, mapping := range strings.Fields(result.StdoutContents()) {
		parts := strings.Split(mapping, ":")
		if len(parts) != 3 {
			continue
		}

		scheme := parts[0]
		host, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		container, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		portMappings[mapping] = PortMapping{
			Scheme:    scheme,
			Host:      host,
			Container: container,
		}
	}

	return portMappings
}

// setPorts sets the ports for a given app
func setPorts(appName string, portMappings []PortMapping) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	currentPorts := getPorts(appName)
	newPorts := map[string]PortMapping{}
	for _, portMapping := range portMappings {
		if _, ok := currentPorts[portMapping.String()]; !ok {
			newPorts[portMapping.String()] = portMapping
		}
	}

	if len(newPorts) == 0 {
		state.State = "present"
		return state
	}

	args := []string{
		"--quiet",
		"ports:add",
		appName,
	}

	for _, portMapping := range newPorts {
		args = append(args, portMapping.String())
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// unsetPorts unsets the ports for a given app
func unsetPorts(appName string, portMappings []PortMapping) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	currentPorts := getPorts(appName)
	removedPorts := map[string]PortMapping{}
	for _, portMapping := range portMappings {
		if _, ok := currentPorts[portMapping.String()]; ok {
			removedPorts[portMapping.String()] = portMapping
		}
	}

	if len(removedPorts) == 0 {
		state.State = "absent"
		return state
	}

	args := []string{
		"--quiet",
		"ports:remove",
		appName,
	}

	for _, portMapping := range removedPorts {
		args = append(args, portMapping.String())
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the PortsTask with the task registry
func init() {
	RegisterTask(&PortsTask{})
}
