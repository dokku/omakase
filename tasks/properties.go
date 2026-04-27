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
// fall back to an unconditional set/unset (matches the pre-probe behavior).
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

// planProperty is the shared Plan() implementation for property tasks. It
// probes the current value via getProperty, returns InSync when current
// matches desired, and otherwise embeds an apply closure that runs the
// underlying `dokku <subcommand>` call. ExecutePlan is the only invoker.
//
// When the report subcommand or property flag is missing, the probe error
// is swallowed and the apply closure runs the set/unset unconditionally,
// matching the pre-probe behavior of property tasks.
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
		StatePresent: func() PlanResult {
			if value == "" {
				return PlanResult{
					Status: PlanStatusError,
					Error:  fmt.Errorf("setting a state of 'present' is invalid without a value for 'value'"),
				}
			}

			// Probe; treat probe failure as "drift, must mutate".
			current, probeErr := getProperty(subcommand, app, global, property)
			if probeErr == nil && current == value {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}

			status := PlanStatusModify
			reason := fmt.Sprintf("would set %s on %s", property, target)
			if probeErr != nil {
				reason = fmt.Sprintf("would set %s on %s (probe failed: %v)", property, target, probeErr)
			} else if current == "" {
				status = PlanStatusCreate
				reason = fmt.Sprintf("%s missing on %s", property, target)
			} else {
				reason = fmt.Sprintf("%s drift on %s (was %q)", property, target, current)
			}

			return PlanResult{
				InSync:    false,
				Status:    status,
				Reason:    reason,
				Mutations: []string{fmt.Sprintf("set %s=%s", property, value)},
				apply:     applyPropertySet(subcommand, target, property, value),
			}
		},
		StateAbsent: func() PlanResult {
			if value != "" {
				return PlanResult{
					Status: PlanStatusError,
					Error:  fmt.Errorf("setting a state of 'absent' is invalid with a value for 'value'"),
				}
			}

			current, probeErr := getProperty(subcommand, app, global, property)
			if probeErr == nil && current == "" {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}

			reason := fmt.Sprintf("would unset %s on %s", property, target)
			if probeErr != nil {
				reason = fmt.Sprintf("would unset %s on %s (probe failed: %v)", property, target, probeErr)
			} else {
				reason = fmt.Sprintf("would unset %s on %s (was %q)", property, target, current)
			}

			return PlanResult{
				InSync:    false,
				Status:    PlanStatusDestroy,
				Reason:    reason,
				Mutations: []string{fmt.Sprintf("unset %s", property)},
				apply:     applyPropertyUnset(subcommand, target, property),
			}
		},
	})
}

// applyPropertySet returns a closure that runs `dokku <subcommand> <target>
// <property> <value>` and converts the result into a TaskOutputState.
func applyPropertySet(subcommand, target, property, value string) func() TaskOutputState {
	return func() TaskOutputState {
		state := TaskOutputState{Changed: false, State: StateAbsent}
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", subcommand, target, property, value},
		})
		state.Commands = append(state.Commands, result.Command)
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
		state.Changed = true
		state.State = StatePresent
		return state
	}
}

// applyPropertyUnset returns a closure that runs `dokku <subcommand> <target>
// <property>` (no value, which dokku interprets as unset) and converts the
// result into a TaskOutputState.
func applyPropertyUnset(subcommand, target, property string) func() TaskOutputState {
	return func() TaskOutputState {
		state := TaskOutputState{Changed: false, State: StatePresent}
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", subcommand, target, property},
		})
		state.Commands = append(state.Commands, result.Command)
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
		state.Changed = true
		state.State = StateAbsent
		return state
	}
}
