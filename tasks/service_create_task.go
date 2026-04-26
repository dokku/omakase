package tasks

import (
	"fmt"
	"docket/subprocess"
)

// ServiceCreateTask creates or destroys a dokku service
type ServiceCreateTask struct {
	// Service is the type of service to create (e.g. redis, postgres, mysql)
	Service string `required:"true" yaml:"service"`

	// Name is the name of the service instance
	Name string `required:"true" yaml:"name"`

	// State is the desired state of the service
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// ServiceCreateTaskExample contains an example of a ServiceCreateTask
type ServiceCreateTaskExample struct {
	// Name is the task name holding the ServiceCreateTask description
	Name string `yaml:"-"`

	// ServiceCreateTask is the ServiceCreateTask configuration
	ServiceCreateTask ServiceCreateTask `yaml:"dokku_service_create"`
}

// GetName returns the name of the example
func (e ServiceCreateTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the service create task
func (t ServiceCreateTask) Doc() string {
	return "Creates or destroys a dokku service"
}

// Examples returns a list of ServiceCreateTaskExamples as yaml
func (t ServiceCreateTask) Examples() ([]Doc, error) {
	return MarshalExamples([]ServiceCreateTaskExample{
		{
			Name: "Create a redis service named my-redis",
			ServiceCreateTask: ServiceCreateTask{
				Service: "redis",
				Name:    "my-redis",
			},
		},
		{
			Name: "Create a postgres service named my-db",
			ServiceCreateTask: ServiceCreateTask{
				Service: "postgres",
				Name:    "my-db",
			},
		},
		{
			Name: "Destroy a redis service named my-redis",
			ServiceCreateTask: ServiceCreateTask{
				Service: "redis",
				Name:    "my-redis",
				State:   "absent",
			},
		},
	})
}

// Execute creates or destroys a dokku service
func (t ServiceCreateTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return createService(t.Service, t.Name) },
		"absent":  func() TaskOutputState { return destroyService(t.Service, t.Name) },
	})
}

// serviceExists checks if a dokku service exists
func serviceExists(service, name string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			fmt.Sprintf("%s:exists", service),
			name,
		},
	})
	if err != nil {
		return false
	}

	return result.ExitCode == 0
}

// createService creates a dokku service
func createService(service, name string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}
	if serviceExists(service, name) {
		state.State = "present"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			fmt.Sprintf("%s:create", service),
			name,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// destroyService destroys a dokku service
func destroyService(service, name string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	if !serviceExists(service, name) {
		state.State = "absent"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"--force",
			fmt.Sprintf("%s:destroy", service),
			name,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the ServiceCreateTask with the task registry
func init() {
	RegisterTask(&ServiceCreateTask{})
}
