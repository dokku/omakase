package subprocess

import (
	"bytes"
)

type CommandResponse struct {
	ExitCode int
	Error    error
	Stderr   []byte
}

func (c CommandResponse) HasError() bool {
	return c.Error != nil
}

func RunDokkuCommand(command []string) CommandResponse {
	var stderr bytes.Buffer
	cmd := NewShellCmdWithArgs("dokku", command...)
	cmd.Command.Stderr = &stderr
	_, err := cmd.Output()
	exitcode := ExitCode(err)

	return CommandResponse{
		ExitCode: exitcode,
		Error:    err,
		Stderr:   stderr.Bytes(),
	}
}
