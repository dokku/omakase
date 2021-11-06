package tasks

import (
	"bytes"
	"fmt"
	"omakase/subprocess"
	"strings"
)

type AppTask struct {
	App   string `required:"true" yaml:"app"`
	State string `required:"true" yaml:"state" default:"present"`
}

func (t AppTask) DesiredState() string {
	return t.State
}

func (t AppTask) Execute() (string, error) {
	command := []string{}
	if t.State == "present" {
		command = []string{"--quiet", "apps:create", t.App}
	} else {
		command = []string{"--quiet", "--force", "apps:destroy", t.App}
	}

	var stderr bytes.Buffer

	cmd := subprocess.NewShellCmdWithArgs("dokku", command...)
	cmd.Command.Stderr = &stderr
	_, err := cmd.Output()

	state := "absent"
	if appExists(t.App) {
		state = "present"
	}

	exitcode := subprocess.ExitCode(err)
	if exitcode == 127 {
		return state, fmt.Errorf("Command not found: dokku")
	}

	if err != nil {
		return state, fmt.Errorf(strings.TrimSpace(stderr.String()))
	}

	return state, nil
}

func (t AppTask) NeedsExecution() bool {
	state := t.State
	if state == "" {
		state = "present"
	}

	exists := appExists(t.App)
	if state == "present" {
		return !exists
	}
	return exists
}

func appExists(appName string) bool {
	cmd := subprocess.NewShellCmdWithArgs("dokku", "--quiet", "apps:exists", appName)
	return cmd.ExecuteQuiet()
}

func init() {
	RegisterTask(&AppTask{})
}
