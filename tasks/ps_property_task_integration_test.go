package tasks

import (
	"testing"
)

func TestIntegrationPsProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-ps-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set ps property
	// Use procfile-path because it goes through dokku's generic property
	// setter and supports unset via the absent state. restart-policy is
	// special-cased on the dokku side and rejects empty values, so it
	// cannot be cleared via `ps:set <app> restart-policy` (no value).
	setTask := PsPropertyTask{
		App:      appName,
		Property: "procfile-path",
		Value:    "Procfile.custom",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set ps property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset ps property
	unsetTask := PsPropertyTask{
		App:      appName,
		Property: "procfile-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset ps property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
