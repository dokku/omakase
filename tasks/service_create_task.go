package tasks

import (
	"fmt"
	"omakase/subprocess"

	yaml "gopkg.in/yaml.v3"
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

// DesiredState returns the desired state of the service
func (t ServiceCreateTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the service create task
func (t ServiceCreateTask) Doc() string {
	return "Creates or destroys a dokku service"
}

// Examples returns a list of ServiceCreateTaskExamples as yaml
func (t ServiceCreateTask) Examples() ([]Doc, error) {
	examples := []ServiceCreateTaskExample{
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

// Execute creates or destroys a dokku service
func (t ServiceCreateTask) Execute() TaskOutputState {
	funcMap := map[State]func(string, string) TaskOutputState{
		"present": createService,
		"absent":  destroyService,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn(t.Service, t.Name)
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
		state.Error = err
		state.Message = result.StderrContents()
		return state
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
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the ServiceCreateTask with the task registry
func init() {
	RegisterTask(&ServiceCreateTask{})
}
