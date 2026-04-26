package tasks

import (
	"testing"
)

func TestIntegrationOpenrestyProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-openresty"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set openresty property
	setTask := OpenrestyPropertyTask{
		App:      appName,
		Property: "proxy-read-timeout",
		Value:    "120s",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set openresty property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset openresty property
	unsetTask := OpenrestyPropertyTask{
		App:      appName,
		Property: "proxy-read-timeout",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset openresty property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
