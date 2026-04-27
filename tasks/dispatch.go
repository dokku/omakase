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

// DispatchPlan looks up the given state in the funcMap, calls the matching
// function, and returns an error PlanResult if the state is not found.
// Mirrors DispatchState but for the read-only Plan() path.
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

// ExecutePlan applies a PlanResult to the server. It is the canonical
// implementation of Task.Execute(): each task's Execute body is
// `return ExecutePlan(t.Plan())`. ExecutePlan ensures the existing
// `state.State == state.DesiredState` contract that commands/apply.go
// relies on.
//
// Three branches:
//
//  1. p.Error != nil  - probe failed; return TaskOutputState carrying the
//     error and the desired state. The apply closure is not invoked.
//  2. p.InSync        - no change needed; return State == DesiredState with
//     Changed=false. The apply closure is not invoked.
//  3. otherwise       - invoke p.apply (must be non-nil) and return its
//     TaskOutputState verbatim. apply is responsible for setting Changed
//     and a final State that matches DesiredState on success.
func ExecutePlan(p PlanResult) TaskOutputState {
	if p.Error != nil {
		return TaskOutputState{
			Error:        p.Error,
			Message:      p.Error.Error(),
			DesiredState: p.DesiredState,
			State:        p.DesiredState,
		}
	}
	if p.InSync {
		return TaskOutputState{
			Changed:      false,
			State:        p.DesiredState,
			DesiredState: p.DesiredState,
		}
	}
	if p.apply == nil {
		return TaskOutputState{
			Error:        fmt.Errorf("plan reports drift but no apply function was provided"),
			DesiredState: p.DesiredState,
			State:        p.DesiredState,
		}
	}
	out := p.apply()
	if out.DesiredState == "" {
		out.DesiredState = p.DesiredState
	}
	return out
}
