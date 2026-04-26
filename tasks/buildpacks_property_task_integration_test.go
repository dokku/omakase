package tasks

import (
	"testing"
)

func TestIntegrationBuildpacksProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-buildpacks-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set buildpacks property
	setTask := BuildpacksPropertyTask{
		App:      appName,
		Property: "stack",
		Value:    "gliderlabs/herokuish:latest",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set buildpacks property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset buildpacks property
	unsetTask := BuildpacksPropertyTask{
		App:      appName,
		Property: "stack",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset buildpacks property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
