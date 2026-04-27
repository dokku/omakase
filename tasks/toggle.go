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
// (state: present) position. nil from a probe (or a non-nil error) is
// treated as "drift, must mutate" so we still run the underlying command.
type ToggleProbe func(ctx ToggleContext) (enabled bool, err error)

// planToggle is the shared Plan() implementation for toggle tasks. The
// probe reports whether the underlying plugin is currently in the
// "enabled" position; when probe is nil or fails, planToggle reports drift
// and the apply closure runs the underlying enable/disable command.
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
		StatePresent: func() PlanResult {
			if probe != nil {
				if enabled, err := probe(ctx); err == nil && enabled {
					return PlanResult{InSync: true, Status: PlanStatusOK}
				}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    fmt.Sprintf("would run %s on %s", enableCmd, target),
				Mutations: []string{fmt.Sprintf("%s %s", enableCmd, target)},
				apply:     applyToggle(enableCmd, target, StatePresent),
			}
		},
		StateAbsent: func() PlanResult {
			if probe != nil {
				if enabled, err := probe(ctx); err == nil && !enabled {
					return PlanResult{InSync: true, Status: PlanStatusOK}
				}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    fmt.Sprintf("would run %s on %s", disableCmd, target),
				Mutations: []string{fmt.Sprintf("%s %s", disableCmd, target)},
				apply:     applyToggle(disableCmd, target, StateAbsent),
			}
		},
	})
}

// applyToggle returns a closure that runs `dokku <subcommand> <target>` and
// reports the resulting state.
func applyToggle(subcommand, target string, finalState State) func() TaskOutputState {
	return func() TaskOutputState {
		state := TaskOutputState{Changed: false, State: finalState}
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", subcommand, target},
		})
		state.Commands = append(state.Commands, result.Command)
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
		state.Changed = true
		state.State = finalState
		return state
	}
}
