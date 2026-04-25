package tasks

import (
	"fmt"
	"docket/subprocess"
)

// NetworkTask creates or destroys a Docker network
type NetworkTask struct {
	// Name is the name of the network
	Name string `required:"true" yaml:"name"`

	// State is the state of the network
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// NetworkTaskExample contains an example of a NetworkTask
type NetworkTaskExample struct {
	// Name is the task name holding the NetworkTask description
	Name string `yaml:"-"`

	// DokkuNetwork is the NetworkTask configuration
	DokkuNetwork NetworkTask `yaml:"dokku_network"`
}

// GetName returns the name of the example
func (e NetworkTaskExample) GetName() string {
	return e.Name
}

// DesiredState returns the desired state of the network
func (t NetworkTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the network task
func (t NetworkTask) Doc() string {
	return "Creates or destroys a Docker network"
}

// Examples returns a list of NetworkTaskExamples as yaml
func (t NetworkTask) Examples() ([]Doc, error) {
	return MarshalExamples([]NetworkTaskExample{
		{
			Name: "Create a network named example-network",
			DokkuNetwork: NetworkTask{
				Name: "example-network",
			},
		},
		{
			Name: "Destroy a network named example-network",
			DokkuNetwork: NetworkTask{
				Name:  "example-network",
				State: "absent",
			},
		},
	})
}

// Execute creates or destroys a Docker network
func (t NetworkTask) Execute() TaskOutputState {
	funcMap := map[State]func(string) TaskOutputState{
		"present": createNetwork,
		"absent":  destroyNetwork,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn(t.Name)
}

// networkExists checks if a Docker network exists
func networkExists(name string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"network:exists",
			name,
		},
	})
	if err != nil {
		return false
	}

	return result.ExitCode == 0
}

// createNetwork creates a Docker network
func createNetwork(name string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}
	if networkExists(name) {
		state.State = "present"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"network:create",
			name,
		},
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

// destroyNetwork destroys a Docker network
func destroyNetwork(name string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	if !networkExists(name) {
		state.State = "absent"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"--force",
			"network:destroy",
			name,
		},
	})
	if err != nil {
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the NetworkTask with the task registry
func init() {
	RegisterTask(&NetworkTask{})
}
