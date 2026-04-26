package tasks

import (
	"testing"
)

func TestIntegrationChecksToggle(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-checks"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// enable checks
	enableTask := ChecksToggleTask{App: appName, State: StatePresent}
	result := enableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to enable checks: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// disable checks
	disableTask := ChecksToggleTask{App: appName, State: StateAbsent}
	result = disableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to disable checks: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
