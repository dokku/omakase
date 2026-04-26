package tasks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dokku/docket/subprocess"
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
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return setConfig(t) },
		"absent":  func() TaskOutputState { return unsetConfig(t) },
	})
}

// Plan reports the drift the ConfigTask would produce.
func (t ConfigTask) Plan() PlanResult {
	return DispatchPlan(t.State, map[State]func() PlanResult{
		"present": func() PlanResult { return planSetConfig(t) },
		"absent":  func() PlanResult { return planUnsetConfig(t) },
	})
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

// configKeysToSet returns keys whose desired value differs from the current value.
func configKeysToSet(current, desired map[string]string) []string {
	keys := []string{}
	for k, v := range desired {
		if cur, ok := current[k]; !ok || cur != v {
			keys = append(keys, k)
		}
	}
	return keys
}

// configKeysToUnset returns keys present in desired that exist in current.
func configKeysToUnset(current, desired map[string]string) []string {
	keys := []string{}
	for k := range desired {
		if _, ok := current[k]; ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// planSetConfig reports drift for a present-state config set.
func planSetConfig(t ConfigTask) PlanResult {
	currentConfig, err := getConfig(t)
	if err != nil {
		return PlanResult{Status: PlanStatusError, Error: err}
	}

	keys := configKeysToSet(currentConfig, t.Config)
	if len(keys) == 0 {
		return PlanResult{InSync: true, Status: PlanStatusOK}
	}

	mutations := make([]string, 0, len(keys))
	status := PlanStatusModify
	allNew := true
	for _, k := range keys {
		if _, ok := currentConfig[k]; ok {
			mutations = append(mutations, fmt.Sprintf("set %s (was set)", k))
			allNew = false
		} else {
			mutations = append(mutations, fmt.Sprintf("set %s (new)", k))
		}
	}
	if allNew {
		status = PlanStatusCreate
	}
	return PlanResult{
		InSync:    false,
		Status:    status,
		Reason:    fmt.Sprintf("%d key(s) to set", len(keys)),
		Mutations: mutations,
	}
}

// planUnsetConfig reports drift for an absent-state config unset.
func planUnsetConfig(t ConfigTask) PlanResult {
	currentConfig, err := getConfig(t)
	if err != nil {
		return PlanResult{Status: PlanStatusError, Error: err}
	}

	keys := configKeysToUnset(currentConfig, t.Config)
	if len(keys) == 0 {
		return PlanResult{InSync: true, Status: PlanStatusOK}
	}

	mutations := make([]string, 0, len(keys))
	for _, k := range keys {
		mutations = append(mutations, fmt.Sprintf("unset %s", k))
	}
	return PlanResult{
		InSync:    false,
		Status:    PlanStatusDestroy,
		Reason:    fmt.Sprintf("%d key(s) to unset", len(keys)),
		Mutations: mutations,
	}
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
		return TaskOutputErrorFromExec(state, err, result)
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
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the ConfigTask with the task registry
func init() {
	RegisterTask(&ConfigTask{})
}
