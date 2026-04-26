package tasks

import (
	"errors"
	"fmt"
	"github.com/dokku/docket/subprocess"
)

// PropertyContext represents the context for a property
type PropertyContext struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the property should be applied globally
	Global bool `required:"false" yaml:"global"`

	// Property is the name of the property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value of the property to set
	Value string `required:"false" yaml:"value"`
}

// executeProperty is a shared Execute implementation for property tasks.
func executeProperty(state State, app string, global bool, property, value, subcommand string) TaskOutputState {
	if !global && app == "" {
		return TaskOutputState{
			Error: errors.New("app is required when global is false"),
		}
	}

	ctx := PropertyContext{
		App:      app,
		Global:   global,
		Property: property,
		Value:    value,
	}
	return DispatchState(state, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return setProperty(subcommand, ctx) },
		"absent":  func() TaskOutputState { return unsetProperty(subcommand, ctx) },
	})
}

// planProperty is a shared Plan implementation for property tasks.
//
// Property tasks today do not probe live state before mutating (see the TODOs
// in setProperty/unsetProperty), so the plan result is conservative: it
// reports drift unconditionally and notes the limitation in Reason. A
// follow-up that adds <subcommand>:report probes can replace this with a
// precise comparison.
func planProperty(state State, app string, global bool, property, value, subcommand string) PlanResult {
	if !global && app == "" {
		return PlanResult{
			Status: PlanStatusError,
			Error:  errors.New("app is required when global is false"),
		}
	}
	if global && app != "" {
		return PlanResult{
			Status: PlanStatusError,
			Error:  fmt.Errorf("'app' must not be set when 'global' is set to true"),
		}
	}

	target := app
	if global {
		target = "--global"
	}

	return DispatchPlan(state, map[State]func() PlanResult{
		"present": func() PlanResult {
			if value == "" {
				return PlanResult{
					Status: PlanStatusError,
					Error:  fmt.Errorf("setting a state of 'present' is invalid without a value for 'value'"),
				}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    fmt.Sprintf("would set %s on %s (current value not probed)", property, target),
				Mutations: []string{fmt.Sprintf("set %s=%s", property, value)},
			}
		},
		"absent": func() PlanResult {
			if value != "" {
				return PlanResult{
					Status: PlanStatusError,
					Error:  fmt.Errorf("setting a state of 'absent' is invalid with a value for 'value'"),
				}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    fmt.Sprintf("would unset %s on %s (current value not probed)", property, target),
				Mutations: []string{fmt.Sprintf("unset %s", property)},
			}
		},
	})
}

// setProperty sets a property for a given app
func setProperty(subcommand string, pctx PropertyContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	if pctx.Global && pctx.App != "" {
		state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
		return state
	}

	appName := "--global"
	if pctx.App != "" {
		appName = pctx.App
	}

	if pctx.Value == "" {
		state.Error = fmt.Errorf("setting a state of 'present' is invalid without a value for 'value'")
		return state
	}

	// todo: validate that the value isn't already set

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			subcommand,
			appName,
			pctx.Property,
			pctx.Value,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// unsetProperty unsets a property for a given app
func unsetProperty(subcommand string, pctx PropertyContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	if pctx.Global && pctx.App != "" {
		state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
		return state
	}

	appName := "--global"
	if pctx.App != "" {
		appName = pctx.App
	}

	if pctx.Value != "" {
		state.Error = fmt.Errorf("setting a state of 'absent' is invalid with a value for 'value'")
		return state
	}

	// todo: validate that the value isn't already unset

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			subcommand,
			appName,
			pctx.Property,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}
