package tasks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"omakase/subprocess"
)

type ConfigTask struct {
	App     string            `required:"true" yaml:"app"`
	Restart bool              `yaml:"restart" default:"true"`
	Config  map[string]string `yaml:"config"`
	State   string            `required:"true" yaml:"state" default:"present"`
}

func (t ConfigTask) DesiredState() string {
	return t.State
}

func (t ConfigTask) Execute() TaskOutputState {
	funcMap := map[string]func(ConfigTask) TaskOutputState{
		"present": setConfig,
		"absent":  unsetConfig,
	}

	fn := funcMap[t.State]
	return fn(t)
}

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
		t.App,
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

func init() {
	RegisterTask(&ConfigTask{})
}
