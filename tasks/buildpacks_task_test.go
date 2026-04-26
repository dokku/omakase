package tasks

import (
	"strings"
	"testing"
)

func TestBuildpacksTaskInvalidState(t *testing.T) {
	task := BuildpacksTask{App: "test-app", Buildpacks: []string{"https://example.com/bp.git"}, State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestBuildpacksTaskPresentMissingApp(t *testing.T) {
	task := BuildpacksTask{Buildpacks: []string{"https://example.com/bp.git"}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'app' is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestBuildpacksTaskAbsentMissingApp(t *testing.T) {
	task := BuildpacksTask{State: StateAbsent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'app' is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestBuildpacksTaskPresentEmptyBuildpacks(t *testing.T) {
	task := BuildpacksTask{App: "test-app", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with empty buildpacks and state=present should return an error")
	}
	if !strings.Contains(result.Error.Error(), "must not be empty") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGetTasksBuildpacksTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: add buildpacks
      dokku_buildpacks:
        app: test-app
        buildpacks:
          - https://github.com/heroku/heroku-buildpack-nodejs.git
          - https://github.com/heroku/heroku-buildpack-nginx.git
        state: present
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("add buildpacks")
	if task == nil {
		t.Fatal("task 'add buildpacks' not found")
	}

	bpTask, ok := task.(*BuildpacksTask)
	if !ok {
		t.Fatalf("task is not a BuildpacksTask (type is %T)", task)
	}
	if bpTask.App != "test-app" {
		t.Errorf("App = %q, want %q", bpTask.App, "test-app")
	}
	if len(bpTask.Buildpacks) != 2 {
		t.Fatalf("expected 2 buildpacks, got %d", len(bpTask.Buildpacks))
	}
	if bpTask.State != StatePresent {
		t.Errorf("State = %q, want %q", bpTask.State, StatePresent)
	}
}
