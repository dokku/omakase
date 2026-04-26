package tasks

import (
	"testing"
)

func TestIntegrationBuilderNixpacksProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder-nixpacks"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder-nixpacks property
	setTask := BuilderNixpacksPropertyTask{
		App:      appName,
		Property: "nixpackstoml-path",
		Value:    "config/nixpacks.toml",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder-nixpacks property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder-nixpacks property
	unsetTask := BuilderNixpacksPropertyTask{
		App:      appName,
		Property: "nixpackstoml-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder-nixpacks property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
