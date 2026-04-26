package tasks

import (
	"testing"
)

func TestIntegrationLogsProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-logs-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set logs property
	setTask := LogsPropertyTask{
		App:      appName,
		Property: "max-size",
		Value:    "100m",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set logs property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset logs property
	unsetTask := LogsPropertyTask{
		App:      appName,
		Property: "max-size",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset logs property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
