package tasks

import (
	"testing"
)

func TestIntegrationAppJsonProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-app-json-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set app.json property
	setTask := AppJsonPropertyTask{
		App:      appName,
		Property: "appjson-path",
		Value:    "app.json",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set app.json property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset app.json property
	unsetTask := AppJsonPropertyTask{
		App:      appName,
		Property: "appjson-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset app.json property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
