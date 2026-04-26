package tasks

import (
	"testing"
)

func TestIntegrationDockerOptions(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-docker-options"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	option := "-v /tmp/docket-test:/tmp/docket-test"
	phase := "deploy"

	// initial state - option not present
	current, err := getDockerOptions(appName)
	if err != nil {
		t.Fatalf("getDockerOptions failed: %v", err)
	}
	if optionPresent(current[phase], option) {
		t.Fatalf("expected option not to be present initially")
	}

	// add option
	addTask := DockerOptionsTask{App: appName, Phase: phase, Option: option, State: StatePresent}
	result := addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to add option: %v", result.Error)
	}
	if !result.Changed {
		t.Errorf("expected Changed=true on first add")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	current, err = getDockerOptions(appName)
	if err != nil {
		t.Fatalf("getDockerOptions failed: %v", err)
	}
	if !optionPresent(current[phase], option) {
		t.Errorf("expected option to be present after add (got %q for phase %s)", current[phase], phase)
	}

	// add same option - idempotent
	result = addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second add: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent add")
	}

	// remove option
	removeTask := DockerOptionsTask{App: appName, Phase: phase, Option: option, State: StateAbsent}
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to remove option: %v", result.Error)
	}
	if !result.Changed {
		t.Errorf("expected Changed=true on first remove")
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	current, err = getDockerOptions(appName)
	if err != nil {
		t.Fatalf("getDockerOptions failed: %v", err)
	}
	if optionPresent(current[phase], option) {
		t.Errorf("expected option not to be present after remove (got %q for phase %s)", current[phase], phase)
	}

	// remove again - idempotent
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second remove: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent remove")
	}
}
