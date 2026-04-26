package tasks

import (
	"testing"
)

func TestIntegrationBuilderRailpackProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder-railpack"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder-railpack property
	setTask := BuilderRailpackPropertyTask{
		App:      appName,
		Property: "railpackjson-path",
		Value:    "config/railpack.json",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder-railpack property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder-railpack property
	unsetTask := BuilderRailpackPropertyTask{
		App:      appName,
		Property: "railpackjson-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder-railpack property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
