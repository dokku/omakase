package tasks

import (
	"testing"
)

func TestIntegrationSchedulerProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-scheduler"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set scheduler property
	setTask := SchedulerPropertyTask{
		App:      appName,
		Property: "selected",
		Value:    "docker-local",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set scheduler property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset scheduler property
	unsetTask := SchedulerPropertyTask{
		App:      appName,
		Property: "selected",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset scheduler property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
