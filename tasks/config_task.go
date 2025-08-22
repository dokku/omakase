package tasks

import (
	"encoding/base64"
	"fmt"
	"omakase/subprocess"
)

type ConfigTask struct {
	App   string `required:"true" yaml:"app"`
	Key   string `required:"true" yaml:"key"`
	Value string `required:"true" yaml:"value"`
	State string `required:"true" yaml:"state" default:"present"`
}

func (t ConfigTask) DesiredState() string {
	return t.State
}

func (t ConfigTask) Execute() TaskOutputState {
	funcMap := map[string]func(string, string, string) TaskOutputState{
		"present": setConfig,
		"absent":  unsetConfig,
	}

	fn := funcMap[t.State]
	return fn(t.App, t.Key, t.Value)
}

func getConfig(app string) (string, bool) {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"config:get",
			app,
		},
		WorkingDirectory: "/tmp",
	})
	if err != nil {
		return "", false
	}
	return result.StdoutContents(), true
}

func setConfig(app, key, value string) TaskOutputState {
	currentValue, ok := getConfig(app)
	if ok && currentValue == value {
		return TaskOutputState{
			Changed: false,
			State:   "present",
		}
	}

	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	// todo: rename no-restart to skip-deploy
	base64Value := base64.StdEncoding.EncodeToString([]byte(value))
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"config:set",
			"--encoded",
			"--no-restart",
			app,
			fmt.Sprintf("%s=%s", key, base64Value),
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

func unsetConfig(app, key, value string) TaskOutputState {
	if _, ok := getConfig(app); !ok {
		return TaskOutputState{
			Changed: false,
			State:   "absent",
		}
	}

	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"config:unset",
			"--no-restart",
			app,
			key,
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

func init() {
	RegisterTask(&ConfigTask{})
}
