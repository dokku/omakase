package tasks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"docket/subprocess"
)

// ConfigTask manages the configuration for a given dokku application
type ConfigTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Restart is a flag indicating if the app should be restarted
	Restart bool `yaml:"restart" default:"true"`

	// Config is a map of configuration key-value pairs
	Config map[string]string `yaml:"config"`

	// State is the desired state of the configuration
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// ConfigTaskExample contains an example of a ConfigTask
type ConfigTaskExample struct {
	// Name is the task name holding the ConfigTask description
	Name string `yaml:"-"`

	// ConfigTask is the ConfigTask configuration
	ConfigTask ConfigTask `yaml:"dokku_config"`
}

// GetName returns the name of the example
func (e ConfigTaskExample) GetName() string {
	return e.Name
}

// DesiredState returns the desired state of the configuration
func (t ConfigTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the config task
func (t ConfigTask) Doc() string {
	return "Manages the configuration for a given dokku application"
}

// Examples returns the examples for the config task
func (t ConfigTask) Examples() ([]Doc, error) {
	return MarshalExamples([]ConfigTaskExample{
		{
			Name: "set KEY=VALUE",
			ConfigTask: ConfigTask{
				App:     "hello-world",
				Restart: true,
				Config: map[string]string{
					"KEY": "VALUE_1",
				},
			},
		},
		{
			Name: "set KEY=VALUE without restart",
			ConfigTask: ConfigTask{
				App:     "hello-world",
				Restart: false,
				Config: map[string]string{
					"KEY": "VALUE_1",
				},
			},
		},
	})
}

// Execute sets or unsets the configuration for a given dokku application
func (t ConfigTask) Execute() TaskOutputState {
	funcMap := map[State]func(ConfigTask) TaskOutputState{
		"present": setConfig,
		"absent":  unsetConfig,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn(t)
}

// getConfig retrieves the current configuration for a given dokku application
func getConfig(t ConfigTask) (map[string]string, error) {
	var config map[string]string
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"config:export",
			"--format",
			"json",
			t.App,
		},
		WorkingDirectory: "/tmp",
	})
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(result.StdoutBytes(), &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// setConfig sets the configuration for a given dokku application
func setConfig(t ConfigTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	currentConfig, err := getConfig(t)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	desiredConfig := make(map[string]string)
	for key, value := range t.Config {
		if _, ok := currentConfig[key]; !ok {
			desiredConfig[key] = value
			continue
		}
		if currentConfig[key] != value {
			desiredConfig[key] = value
		}
	}

	if len(desiredConfig) == 0 {
		state.State = "present"
		return state
	}

	args := []string{
		"--quiet",
		"config:set",
		"--encoded",
	}

	if !t.Restart {
		// todo: rename no-restart to skip-deploy in dokku
		args = append(args, "--no-restart")
	}

	args = append(args, t.App)

	for key, value := range desiredConfig {
		args = append(args, fmt.Sprintf("%s=%s", key, base64.StdEncoding.EncodeToString([]byte(value))))
	}

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

// unsetConfig unsets the configuration for a given dokku application
func unsetConfig(t ConfigTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	currentConfig, err := getConfig(t)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	desiredConfig := make(map[string]bool)
	for key := range t.Config {
		if _, ok := currentConfig[key]; ok {
			desiredConfig[key] = true
		}
	}

	if len(desiredConfig) == 0 {
		state.State = "absent"
		return state
	}

	args := []string{
		"--quiet",
		"config:unset",
	}

	if !t.Restart {
		// todo: rename no-restart to skip-deploy in dokku
		args = append(args, "--no-restart")
	}

	args = append(args, t.App)

	for key := range desiredConfig {
		args = append(args, key)
	}

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
	state.State = "absent"
	return state
}

// init registers the ConfigTask with the task registry
func init() {
	RegisterTask(&ConfigTask{})
}
