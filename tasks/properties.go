package tasks

import (
	"errors"
	"fmt"
	"strings"

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

// planProperty is a shared Plan implementation for property tasks. It probes
// the current value via `dokku <plugin>:report <target> --<plugin>-<property>`
// and compares against the desired value. When the probe fails (plugin
// without :report support or unknown property flag), falls back to reporting
// drift conservatively rather than erroring, so users still see a usable plan.
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
			current, err := getProperty(subcommand, app, global, property)
			if err != nil {
				return PlanResult{
					InSync:    false,
					Status:    PlanStatusModify,
					Reason:    fmt.Sprintf("would set %s on %s (probe failed: %v)", property, target, err),
					Mutations: []string{fmt.Sprintf("set %s=%s", property, value)},
				}
			}
			if current == value {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			status := PlanStatusModify
			if current == "" {
				status = PlanStatusCreate
			}
			reason := fmt.Sprintf("%s drift on %s", property, target)
			if current != "" {
				reason = fmt.Sprintf("%s drift on %s (was %q)", property, target, current)
			}
			return PlanResult{
				InSync:    false,
				Status:    status,
				Reason:    reason,
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
			current, err := getProperty(subcommand, app, global, property)
			if err != nil {
				return PlanResult{
					InSync:    false,
					Status:    PlanStatusModify,
					Reason:    fmt.Sprintf("would unset %s on %s (probe failed: %v)", property, target, err),
					Mutations: []string{fmt.Sprintf("unset %s", property)},
				}
			}
			if current == "" {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusDestroy,
				Reason:    fmt.Sprintf("would unset %s on %s (was %q)", property, target, current),
				Mutations: []string{fmt.Sprintf("unset %s", property)},
			}
		},
	})
}

// pluginFromSubcommand returns the plugin component of a colon-separated
// subcommand. For example, "nginx:set" -> "nginx", "buildpacks:set-property" ->
// "buildpacks", "app-json:set" -> "app-json".
func pluginFromSubcommand(subcommand string) string {
	return strings.SplitN(subcommand, ":", 2)[0]
}

// getProperty reads the current value of a property via
// `dokku <plugin>:report <target> --<plugin>-<property>`. Returns the
// trimmed string value on success. When the report subcommand or property
// flag is not supported by the plugin, returns a non-nil error and callers
// fall back to a conservative path (Plan reports drift; Execute proceeds
// to set/unset unconditionally).
func getProperty(subcommand, app string, global bool, property string) (string, error) {
	plugin := pluginFromSubcommand(subcommand)
	reportSubcommand := plugin + ":report"
	reportFlag := "--" + plugin + "-" + property

	args := []string{"--quiet", reportSubcommand}
	if global {
		args = append(args, "--global", reportFlag)
	} else {
		args = append(args, app, reportFlag)
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.StdoutContents()), nil
}

// setProperty sets a property for a given app, short-circuiting when the
// current value already matches the desired value.
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

	// Probe current value; if it matches desired, no-op (Changed=false).
	// A failed probe (plugin without :report or unsupported property flag)
	// falls through to the unconditional set, matching pre-probe behavior.
	if current, err := getProperty(subcommand, pctx.App, pctx.Global, pctx.Property); err == nil && current == pctx.Value {
		state.State = "present"
		return state
	}

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

// unsetProperty unsets a property for a given app, short-circuiting when the
// current value is already empty.
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

	if current, err := getProperty(subcommand, pctx.App, pctx.Global, pctx.Property); err == nil && current == "" {
		state.State = "absent"
		return state
	}

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
