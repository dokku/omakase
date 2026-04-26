package tasks

import (
	"testing"
)

func TestGitSyncTaskInvalidState(t *testing.T) {
	task := GitSyncTask{
		App:    "test-app",
		Remote: "https://example.com/repo",
		State:  "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}
