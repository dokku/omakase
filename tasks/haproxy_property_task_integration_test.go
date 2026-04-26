package tasks

import (
	"testing"
)

func TestIntegrationHaproxyProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-haproxy"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set haproxy property
	setTask := HaproxyPropertyTask{
		App:      appName,
		Property: "letsencrypt-email",
		Value:    "admin@example.com",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set haproxy property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset haproxy property
	unsetTask := HaproxyPropertyTask{
		App:      appName,
		Property: "letsencrypt-email",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset haproxy property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
