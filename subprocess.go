package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ShellCmd represents a shell command to be run for dokku
type ShellCmd struct {
	Env           map[string]string
	Command       *exec.Cmd
	CommandString string
	Args          []string
	ShowOutput    bool
}

// NewShellCmdWithArgs returns a new ShellCmd struct
func NewShellCmdWithArgs(cmd string, args ...string) *ShellCmd {
	commandString := strings.Join(append([]string{cmd}, args...), " ")

	return &ShellCmd{
		Command:       exec.Command(cmd, args...),
		CommandString: commandString,
		Args:          args,
		ShowOutput:    true,
	}
}

func (sc *ShellCmd) setup() {
	env := os.Environ()
	for k, v := range sc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	sc.Command.Env = env
	if sc.ShowOutput {
		sc.Command.Stdout = os.Stdout
		sc.Command.Stderr = os.Stderr
	}
}

// Execute is a lightweight wrapper around exec.Command
func (sc *ShellCmd) Execute() bool {
	sc.setup()

	if err := sc.Command.Run(); err != nil {
		return false
	}
	return true
}

// Execute is a lightweight wrapper around exec.Command
func (sc *ShellCmd) ExecuteQuiet() bool {
	sc.ShowOutput = false
	sc.setup()

	if err := sc.Command.Run(); err != nil {
		return false
	}
	return true
}

// Start is a wrapper around exec.Command.Start()
func (sc *ShellCmd) Start() error {
	sc.setup()

	return sc.Command.Start()
}

// Output is a lightweight wrapper around exec.Command.Output()
func (sc *ShellCmd) Output() ([]byte, error) {
	sc.ShowOutput = false
	sc.setup()
	return sc.Command.Output()
}

// CombinedOutput is a lightweight wrapper around exec.Command.CombinedOutput()
func (sc *ShellCmd) CombinedOutput() ([]byte, error) {
	sc.ShowOutput = false
	sc.setup()
	return sc.Command.CombinedOutput()
}

func exitCode(err error) int {
	exitcode := 0
	if err != nil {
		exitcode = 1

		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitcode = status.ExitStatus()
			}
		}
	}

	return exitcode
}
