package tasks

import (
	"testing"
)

func TestIntegrationBuilderPackProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder-pack"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder-pack property
	setTask := BuilderPackPropertyTask{
		App:      appName,
		Property: "projecttoml-path",
		Value:    "config/project.toml",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder-pack property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder-pack property
	unsetTask := BuilderPackPropertyTask{
		App:      appName,
		Property: "projecttoml-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder-pack property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
