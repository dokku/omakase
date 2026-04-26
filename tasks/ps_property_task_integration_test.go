package tasks

import (
	"testing"
)

func TestIntegrationPsProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-ps-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set ps property
	setTask := PsPropertyTask{
		App:      appName,
		Property: "restart-policy",
		Value:    "on-failure:5",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set ps property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset ps property
	unsetTask := PsPropertyTask{
		App:      appName,
		Property: "restart-policy",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset ps property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
