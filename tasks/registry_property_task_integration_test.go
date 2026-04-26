package tasks

import (
	"testing"
)

func TestIntegrationRegistryProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-registry"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set registry property
	setTask := RegistryPropertyTask{
		App:      appName,
		Property: "image-repo",
		Value:    "registry.example.com/" + appName,
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set registry property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset registry property
	unsetTask := RegistryPropertyTask{
		App:      appName,
		Property: "image-repo",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset registry property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
