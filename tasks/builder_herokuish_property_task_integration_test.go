package tasks

import (
	"testing"
)

func TestIntegrationBuilderHerokuishProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder-herokuish"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder-herokuish property
	setTask := BuilderHerokuishPropertyTask{
		App:      appName,
		Property: "allowed",
		Value:    "true",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder-herokuish property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder-herokuish property
	unsetTask := BuilderHerokuishPropertyTask{
		App:      appName,
		Property: "allowed",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder-herokuish property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
