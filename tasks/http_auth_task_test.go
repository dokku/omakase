package tasks

import (
	"strings"
	"testing"

	_ "github.com/gliderlabs/sigil/builtin"
)

func TestHttpAuthTaskInvalidState(t *testing.T) {
	task := HttpAuthTask{App: "test-app", Username: "admin", Password: "secret", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestHttpAuthTaskPresentWithoutUsername(t *testing.T) {
	task := HttpAuthTask{App: "test-app", Password: "secret", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no username")
	}
	if !strings.Contains(result.Error.Error(), "username is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestHttpAuthTaskPresentWithoutPassword(t *testing.T) {
	task := HttpAuthTask{App: "test-app", Username: "admin", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no password")
	}
	if !strings.Contains(result.Error.Error(), "password is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGetTasksHttpAuthTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: enable http auth
      dokku_http_auth:
        app: test-app
        username: admin
        password: secret
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("enable http auth")
	if task == nil {
		t.Fatal("task 'enable http auth' not found")
	}

	haTask, ok := task.(*HttpAuthTask)
	if !ok {
		ht, ok2 := task.(HttpAuthTask)
		if !ok2 {
			t.Fatalf("task is not an HttpAuthTask (type is %T)", task)
		}
		haTask = &ht
	}

	if haTask.App != "test-app" {
		t.Errorf("App = %q, want %q", haTask.App, "test-app")
	}
	if haTask.Username != "admin" {
		t.Errorf("Username = %q, want %q", haTask.Username, "admin")
	}
	if haTask.Password != "secret" {
		t.Errorf("Password = %q, want %q", haTask.Password, "secret")
	}
	if haTask.State != StatePresent {
		t.Errorf("expected default state 'present', got %q", haTask.State)
	}
}
