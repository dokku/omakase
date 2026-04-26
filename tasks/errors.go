package tasks

import "github.com/dokku/docket/subprocess"

// TaskOutputErrorFromExec sets Error and Message on the given state from an exec result.
// Returns the modified state for convenient use with early returns.
func TaskOutputErrorFromExec(state TaskOutputState, err error, result subprocess.ExecCommandResponse) TaskOutputState {
	state.Error = err
	state.Message = result.StderrContents()
	return state
}
