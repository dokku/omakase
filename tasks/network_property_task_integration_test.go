package tasks

import (
	"testing"
)

func TestIntegrationNetworkProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-network"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set network property
	setTask := NetworkPropertyTask{
		App:      appName,
		Property: "bind-all-interfaces",
		Value:    "true",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set network property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset network property
	unsetTask := NetworkPropertyTask{
		App:      appName,
		Property: "bind-all-interfaces",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset network property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
