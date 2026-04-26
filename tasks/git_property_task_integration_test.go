package tasks

import (
	"testing"
)

func TestIntegrationGitProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-git-prop"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set git property
	setTask := GitPropertyTask{
		App:      appName,
		Property: "deploy-branch",
		Value:    "main",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set git property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset git property
	unsetTask := GitPropertyTask{
		App:      appName,
		Property: "deploy-branch",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset git property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
