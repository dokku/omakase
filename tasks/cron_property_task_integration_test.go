package tasks

import (
	"testing"
)

func TestIntegrationCronProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-cron"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set cron property
	setTask := CronPropertyTask{
		App:      appName,
		Property: "maintenance",
		Value:    "true",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set cron property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset cron property
	unsetTask := CronPropertyTask{
		App:      appName,
		Property: "maintenance",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset cron property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
