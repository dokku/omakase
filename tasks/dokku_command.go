package tasks

import (
	"bytes"
	"omakase/subprocess"
)

type CommandResponse struct {
	ExitCode int
	Error    error
	Stderr   []byte
}

func (c CommandResponse) HasError() bool {
	return c.Error != nil
}

func runDokkuCommand(command []string) CommandResponse {
	var stderr bytes.Buffer
	cmd := subprocess.NewShellCmdWithArgs("dokku", command...)
	cmd.Command.Stderr = &stderr
	_, err := cmd.Output()
	exitcode := subprocess.ExitCode(err)

	return CommandResponse{
		ExitCode: exitcode,
		Error:    err,
		Stderr:   stderr.Bytes(),
	}
}
