package tasks

import (
	"encoding/json"
	"fmt"
	"omakase/subprocess"
)

// StorageMountTask manages the storage for a given dokku application
type StorageMountTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// HostDir is the host directory to mount
	HostDir string `required:"true" yaml:"host_dir"`

	// ContainerDir is the container directory to mount
	ContainerDir string `required:"true" yaml:"container_dir"`

	// State is the desired state of the storage
	State State `required:"false" yaml:"state" default:"present" options:"present,absent"`
}

// DesiredState returns the desired state of the storage
func (t StorageMountTask) DesiredState() State {
	return t.State
}

// Execute mounts or unmounts the storage for a given app
func (t StorageMountTask) Execute() TaskOutputState {
	funcMap := map[State]func(string, string, string) TaskOutputState{
		"present": mountStorage,
		"absent":  unmountStorage,
	}

	fn := funcMap[t.State]
	return fn(t.App, t.HostDir, t.ContainerDir)
}

// mountExists checks if the storage mount exists
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
		HostPath      string `json:"host_path"`
		ContainerPath string `json:"container_path"`
	}

	err = json.Unmarshal(result.StdoutBytes(), &mounts)
	if err != nil {
		return false
	}

	for _, mount := range mounts {
		if mount.HostPath == hostDir && mount.ContainerPath == containerDir {
			return true
		}
	}
	return false
}

// mountStorage mounts the storage for a given app
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

// unmountStorage unmounts the storage for a given app
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

// init registers the StorageMountTask with the task registry
func init() {
	RegisterTask(&StorageMountTask{})
}
