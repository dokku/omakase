package tasks

import (
	"testing"
)

func TestIntegrationNginxProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-nginx"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set nginx property
	setTask := NginxPropertyTask{
		App:      appName,
		Property: "proxy-read-timeout",
		Value:    "120s",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set nginx property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset nginx property
	unsetTask := NginxPropertyTask{
		App:      appName,
		Property: "proxy-read-timeout",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset nginx property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
