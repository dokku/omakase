package tasks

import (
	"testing"
)

func TestIntegrationSchedulerK3sProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-scheduler-k3s"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set scheduler-k3s property
	setTask := SchedulerK3sPropertyTask{
		App:      appName,
		Property: "deploy-timeout",
		Value:    "300s",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set scheduler-k3s property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset scheduler-k3s property
	unsetTask := SchedulerK3sPropertyTask{
		App:      appName,
		Property: "deploy-timeout",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset scheduler-k3s property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
