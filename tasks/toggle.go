package tasks

import (
	"fmt"

	"github.com/dokku/docket/subprocess"
)

// ToggleContext represents the context for a toggle operation
type ToggleContext struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Global is a flag indicating if the toggle should be applied globally
	Global bool `required:"false" yaml:"global"`

	// AllowGlobal is a flag indicating if the toggle should be applied globally
	AllowGlobal bool `required:"false" yaml:"allow_global"`
}

// ToggleProbe returns whether the toggle is currently in the "enabled"
// (state: present) position. It is invoked by both Plan and Execute. When
// probe fails (plugin missing the corresponding report flag, transient
// dokku error), the second return value is non-nil and callers fall back to
// reporting drift / running the underlying enable / disable command.
type ToggleProbe func(ctx ToggleContext) (enabled bool, err error)

// enablePlugin executes the enable state for a plugin
func enablePlugin(subcommand string, pctx ToggleContext, probe ToggleProbe) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	appName := pctx.App
	if pctx.AllowGlobal {
		if pctx.Global && pctx.App != "" {
			state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
			return state
		}
		if pctx.Global {
			appName = "--global"
		}
	}

	if probe != nil {
		if enabled, err := probe(pctx); err == nil && enabled {
			state.State = "present"
			return state
		}
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			subcommand,
			appName,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// executeToggle is a shared Execute implementation for toggle tasks. The
// probe function reports whether the underlying plugin is currently in the
// "enabled" position; when nil, both states run unconditionally (matches the
// pre-probe behavior).
func executeToggle(state State, app string, global bool, allowGlobal bool, enableCmd, disableCmd string, probe ToggleProbe) TaskOutputState {
	ctx := ToggleContext{
		AllowGlobal: allowGlobal,
		App:         app,
		Global:      global,
	}
	return DispatchState(state, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return enablePlugin(enableCmd, ctx, probe) },
		"absent":  func() TaskOutputState { return disablePlugin(disableCmd, ctx, probe) },
	})
}

// planToggle is a shared Plan implementation for toggle tasks. probe is
// called to determine the current enabled state. When probe fails, the
// plan reports drift conservatively and notes the limitation in Reason.
func planToggle(state State, app string, global bool, allowGlobal bool, enableCmd, disableCmd string, probe ToggleProbe) PlanResult {
	if allowGlobal && global && app != "" {
		return PlanResult{
			Status: PlanStatusError,
			Error:  fmt.Errorf("'app' must not be set when 'global' is set to true"),
		}
	}

	target := app
	if allowGlobal && global {
		target = "--global"
	}

	ctx := ToggleContext{
		AllowGlobal: allowGlobal,
		App:         app,
		Global:      global,
	}

	return DispatchPlan(state, map[State]func() PlanResult{
		"present": func() PlanResult {
			if probe == nil {
				return PlanResult{
					InSync:    false,
					Status:    PlanStatusModify,
					Reason:    fmt.Sprintf("would run %s on %s (no probe)", enableCmd, target),
					Mutations: []string{fmt.Sprintf("%s %s", enableCmd, target)},
				}
			}
			enabled, err := probe(ctx)
			if err != nil {
				return PlanResult{
					InSync:    false,
					Status:    PlanStatusModify,
					Reason:    fmt.Sprintf("would run %s on %s (probe failed: %v)", enableCmd, target, err),
					Mutations: []string{fmt.Sprintf("%s %s", enableCmd, target)},
				}
			}
			if enabled {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    fmt.Sprintf("disabled on %s", target),
				Mutations: []string{fmt.Sprintf("%s %s", enableCmd, target)},
			}
		},
		"absent": func() PlanResult {
			if probe == nil {
				return PlanResult{
					InSync:    false,
					Status:    PlanStatusModify,
					Reason:    fmt.Sprintf("would run %s on %s (no probe)", disableCmd, target),
					Mutations: []string{fmt.Sprintf("%s %s", disableCmd, target)},
				}
			}
			enabled, err := probe(ctx)
			if err != nil {
				return PlanResult{
					InSync:    false,
					Status:    PlanStatusModify,
					Reason:    fmt.Sprintf("would run %s on %s (probe failed: %v)", disableCmd, target, err),
					Mutations: []string{fmt.Sprintf("%s %s", disableCmd, target)},
				}
			}
			if !enabled {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    fmt.Sprintf("enabled on %s", target),
				Mutations: []string{fmt.Sprintf("%s %s", disableCmd, target)},
			}
		},
	})
}

// disablePlugin executes the disable state for a plugin
func disablePlugin(subcommand string, pctx ToggleContext, probe ToggleProbe) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	appName := pctx.App
	if pctx.AllowGlobal {
		if pctx.Global && pctx.App != "" {
			state.Error = fmt.Errorf("'app' must not be set when 'global' is set to true")
			return state
		}
		if pctx.Global {
			appName = "--global"
		}
	}

	if probe != nil {
		if enabled, err := probe(pctx); err == nil && !enabled {
			state.State = "absent"
			return state
		}
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			subcommand,
			appName,
		},
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}
