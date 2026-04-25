package tasks

import (
	"errors"
	"fmt"
	"docket/subprocess"
	"strings"
)

// ResourceContext represents the context for a resource operation
type ResourceContext struct {
	// App is the name of the app
	App string

	// ProcessType is the process type to filter resources for
	ProcessType string

	// Resources is a map of resource type to quantity
	Resources map[string]string

	// ClearBefore clears all resources before applying new ones
	ClearBefore bool
}

// executeResource is a shared Execute implementation for resource tasks.
func executeResource(state State, app, processType string, resources map[string]string, clearBefore bool, subcommand string) TaskOutputState {
	if state == StatePresent && len(resources) == 0 {
		return TaskOutputState{
			Error:   errors.New("resources are required when state is present"),
			Message: "resources are required when state is present",
		}
	}

	rctx := ResourceContext{
		App:         app,
		ProcessType: processType,
		Resources:   resources,
		ClearBefore: clearBefore,
	}
	return DispatchState(state, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return setResource(subcommand, rctx) },
		"absent":  func() TaskOutputState { return clearResource(subcommand, rctx) },
	})
}

// getResources retrieves the current resources for a given dokku application
func getResources(subcommand string, rctx ResourceContext) (map[string]string, error) {
	args := []string{
		"--quiet",
		subcommand,
	}

	if rctx.ProcessType != "" {
		args = append(args, "--process-type", rctx.ProcessType)
	}

	args = append(args, rctx.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return nil, err
	}

	resources := map[string]string{}
	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			resources[key] = value
		}
	}

	return resources, nil
}

// setResource sets the resources for a given dokku application
func setResource(subcommand string, rctx ResourceContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	if rctx.ClearBefore {
		err := execResourceClear(subcommand, rctx)
		if err != nil {
			state.Error = err
			state.Message = err.Error()
			return state
		}
		state.Changed = true
	}

	if !rctx.ClearBefore {
		currentResources, err := getResources(subcommand, rctx)
		if err != nil {
			state.Error = err
			state.Message = err.Error()
			return state
		}

		// validate that all requested resources exist
		for k := range rctx.Resources {
			if _, ok := currentResources[k]; !ok {
				state.Error = fmt.Errorf("unknown resource %s, valid resources: %v", k, mapKeys(currentResources))
				state.Message = state.Error.Error()
				return state
			}
		}

		hasChanged := false
		for k, v := range rctx.Resources {
			if currentResources[k] != v {
				hasChanged = true
				break
			}
		}

		if !hasChanged {
			state.State = "present"
			return state
		}
	}

	args := []string{
		subcommand,
	}

	for key, value := range rctx.Resources {
		args = append(args, fmt.Sprintf("--%s", key), value)
	}

	if rctx.ProcessType != "" {
		args = append(args, "--process-type", rctx.ProcessType)
	}

	args = append(args, rctx.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// clearResource clears the resources for a given dokku application
func clearResource(subcommand string, rctx ResourceContext) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}

	currentResources, err := getResources(subcommand, rctx)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	hasResources := false
	for _, v := range currentResources {
		if v != "" && v != "0" {
			hasResources = true
			break
		}
	}

	if !hasResources {
		state.State = "absent"
		return state
	}

	err = execResourceClear(subcommand, rctx)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// execResourceClear executes the dokku resource clear command
func execResourceClear(subcommand string, rctx ResourceContext) error {
	args := []string{
		subcommand + "-clear",
	}

	if rctx.ProcessType != "" {
		args = append(args, "--process-type", rctx.ProcessType)
	}

	args = append(args, rctx.App)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return fmt.Errorf("%s", result.StderrContents())
	}

	return nil
}

// mapKeys returns the keys of a map as a slice
func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
