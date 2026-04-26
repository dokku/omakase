package tasks

import (
	"testing"
)

func TestIntegrationGitSync(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-gitsync"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	task := GitSyncTask{
		App:    appName,
		Remote: "https://github.com/dokku/smoke-test-app",
		State:  StatePresent,
	}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to sync git: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for git sync")
	}
}
