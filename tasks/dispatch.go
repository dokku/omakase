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

	result := fn()
	result.DesiredState = state
	return result
}

// DispatchPlan looks up the given state in the funcMap, calls the matching function,
// and returns an error PlanResult if the state is not found. Mirrors DispatchState
// but returns a PlanResult so tasks can implement Plan() with the same shape they
// use for Execute().
func DispatchPlan(state State, funcMap map[State]func() PlanResult) PlanResult {
	fn, ok := funcMap[state]
	if !ok {
		return PlanResult{
			Status: PlanStatusError,
			Error:  fmt.Errorf("invalid state: %s", state),
		}
	}

	result := fn()
	result.DesiredState = state
	return result
}
