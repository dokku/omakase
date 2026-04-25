package tasks

import "fmt"

// DispatchState looks up the given state in the funcMap, calls the matching function,
// and returns an error TaskOutputState if the state is not found.
func DispatchState(state State, funcMap map[State]func() TaskOutputState) TaskOutputState {
	fn, ok := funcMap[state]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", state),
		}
	}

	return fn()
}
