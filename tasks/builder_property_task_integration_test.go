package tasks

import (
	"testing"
)

func TestIntegrationBuilderProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder property
	setTask := BuilderPropertyTask{
		App:      appName,
		Property: "selected",
		Value:    "dockerfile",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder property
	unsetTask := BuilderPropertyTask{
		App:      appName,
		Property: "selected",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
