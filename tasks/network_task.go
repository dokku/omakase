package tasks

import (
	"github.com/dokku/docket/subprocess"
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
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the NetworkTask would produce.
func (t NetworkTask) Plan() PlanResult {
	return DispatchPlan(t.State, map[State]func() PlanResult{
		StatePresent: func() PlanResult {
			if networkExists(t.Name) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusCreate,
				Reason:    "network missing",
				Mutations: []string{"create network " + t.Name},
				apply: func() TaskOutputState {
					state := TaskOutputState{Changed: false, State: StateAbsent}
					result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
						Command: "dokku",
						Args:    []string{"--quiet", "network:create", t.Name},
					})
					if err != nil {
						return TaskOutputErrorFromExec(state, err, result)
					}
					state.Changed = true
					state.State = StatePresent
					return state
				},
			}
		},
		StateAbsent: func() PlanResult {
			if !networkExists(t.Name) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusDestroy,
				Reason:    "network present",
				Mutations: []string{"destroy network " + t.Name},
				apply: func() TaskOutputState {
					state := TaskOutputState{Changed: false, State: StatePresent}
					result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
						Command: "dokku",
						Args:    []string{"--quiet", "--force", "network:destroy", t.Name},
					})
					if err != nil {
						return TaskOutputErrorFromExec(state, err, result)
					}
					state.Changed = true
					state.State = StateAbsent
					return state
				},
			}
		},
	})
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

// destroyNetwork is retained as an integration-test helper. It runs the
// destroy-network apply path synchronously.
func destroyNetwork(name string) TaskOutputState {
	state := TaskOutputState{Changed: false, State: StatePresent}
	if !networkExists(name) {
		state.State = StateAbsent
		return state
	}
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "--force", "network:destroy", name},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}
	state.Changed = true
	state.State = StateAbsent
	return state
}

// init registers the NetworkTask with the task registry
func init() {
	RegisterTask(&NetworkTask{})
}
