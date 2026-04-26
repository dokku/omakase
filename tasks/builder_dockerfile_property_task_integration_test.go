package tasks

import (
	"testing"
)

func TestIntegrationBuilderDockerfileProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-builder-dockerfile"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder-dockerfile property
	setTask := BuilderDockerfilePropertyTask{
		App:      appName,
		Property: "dockerfile-path",
		Value:    "Dockerfile.production",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder-dockerfile property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder-dockerfile property
	unsetTask := BuilderDockerfilePropertyTask{
		App:      appName,
		Property: "dockerfile-path",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder-dockerfile property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
