package tasks

import (
	"testing"
)

func TestIntegrationCaddyProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-caddy"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set caddy property
	setTask := CaddyPropertyTask{
		App:      appName,
		Property: "tls-internal",
		Value:    "true",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set caddy property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset caddy property
	unsetTask := CaddyPropertyTask{
		App:      appName,
		Property: "tls-internal",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset caddy property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
