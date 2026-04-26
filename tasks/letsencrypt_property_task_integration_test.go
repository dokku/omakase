package tasks

import (
	"testing"
)

func TestIntegrationLetsencryptProperty(t *testing.T) {
	skipIfNoDokkuT(t)
	skipIfPluginMissingT(t, "letsencrypt")

	appName := "docket-test-letsencrypt"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set letsencrypt property
	setTask := LetsencryptPropertyTask{
		App:      appName,
		Property: "email",
		Value:    "admin@example.com",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set letsencrypt property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset letsencrypt property
	unsetTask := LetsencryptPropertyTask{
		App:      appName,
		Property: "email",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset letsencrypt property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
