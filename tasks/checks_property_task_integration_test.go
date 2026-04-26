package tasks

import (
	"testing"
)

func TestIntegrationChecksProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-checks-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set checks property
	setTask := ChecksPropertyTask{
		App:      appName,
		Property: "wait-to-retire",
		Value:    "60",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set checks property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset checks property
	unsetTask := ChecksPropertyTask{
		App:      appName,
		Property: "wait-to-retire",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset checks property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
