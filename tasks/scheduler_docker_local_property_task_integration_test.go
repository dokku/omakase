package tasks

import (
	"testing"
)

func TestIntegrationSchedulerDockerLocalProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-scheduler-docker-local"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set scheduler-docker-local property
	setTask := SchedulerDockerLocalPropertyTask{
		App:      appName,
		Property: "init-process",
		Value:    "true",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set scheduler-docker-local property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset scheduler-docker-local property
	unsetTask := SchedulerDockerLocalPropertyTask{
		App:      appName,
		Property: "init-process",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset scheduler-docker-local property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
