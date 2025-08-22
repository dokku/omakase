package tasks

import (
	"encoding/json"
	"fmt"
	"omakase/subprocess"
)

type StorageMountTask struct {
	App          string `required:"true" yaml:"app"`
	HostDir      string `required:"true" yaml:"host_dir"`
	ContainerDir string `required:"true" yaml:"container_dir"`
	State        string `required:"true" yaml:"state" default:"present"`
}

func (t StorageMountTask) DesiredState() string {
	return t.State
}

func (t StorageMountTask) Execute() TaskOutputState {
	funcMap := map[string]func(string, string, string) TaskOutputState{
		"present": mountStorage,
		"absent":  unmountStorage,
	}

	fn := funcMap[t.State]
	return fn(t.App, t.HostDir, t.ContainerDir)
}

func mountExists(app, hostDir, containerDir string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"storage:list",
			app,
			"--format",
			"json",
		},
	})
	if err != nil {
		return false
	}

	var mounts []struct {
		HostDir      string `json:"host_dir"`
		ContainerDir string `json:"container_dir"`
	}

	err = json.Unmarshal(result.StdoutBytes(), &mounts)
	if err != nil {
		return false
	}

	for _, mount := range mounts {
		if mount.HostDir == hostDir && mount.ContainerDir == containerDir {
			return true
		}
	}
	return false
}

func mountStorage(app, hostDir, containerDir string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	// check if the mount already exists
	if mountExists(app, hostDir, containerDir) {
		state.Changed = false
		state.State = "present"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"storage:mount",
			app,
			fmt.Sprintf("%s:%s", hostDir, containerDir),
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

func unmountStorage(app, hostDir, containerDir string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	if !mountExists(app, hostDir, containerDir) {
		state.Changed = false
		state.State = "absent"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"storage:unmount",
			app,
			fmt.Sprintf("%s:%s", hostDir, containerDir),
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
	RegisterTask(&StorageMountTask{})
}
