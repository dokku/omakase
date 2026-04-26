package tasks

import (
	"strings"
	"testing"
)

func TestCertsTaskInvalidState(t *testing.T) {
	task := CertsTask{App: "test-app", Cert: "/tmp/cert", Key: "/tmp/key", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestCertsTaskMissingApp(t *testing.T) {
	task := CertsTask{Cert: "/tmp/cert", Key: "/tmp/key", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'app' is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestCertsTaskGlobalWithApp(t *testing.T) {
	task := CertsTask{App: "test-app", Global: true, Cert: "/tmp/cert", Key: "/tmp/key", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestCertsTaskPresentMissingCert(t *testing.T) {
	task := CertsTask{App: "test-app", Key: "/tmp/key", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without cert should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'cert' and 'key' are required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestCertsTaskPresentMissingKey(t *testing.T) {
	task := CertsTask{App: "test-app", Cert: "/tmp/cert", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without key should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'cert' and 'key' are required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGetTasksCertsTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: install cert
      dokku_certs:
        app: test-app
        cert: /etc/ssl/test-app.crt
        key: /etc/ssl/test-app.key
        state: present
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("install cert")
	if task == nil {
		t.Fatal("task 'install cert' not found")
	}

	certsTask, ok := task.(*CertsTask)
	if !ok {
		t.Fatalf("task is not a CertsTask (type is %T)", task)
	}
	if certsTask.App != "test-app" {
		t.Errorf("App = %q, want %q", certsTask.App, "test-app")
	}
	if certsTask.Cert != "/etc/ssl/test-app.crt" {
		t.Errorf("Cert = %q, want %q", certsTask.Cert, "/etc/ssl/test-app.crt")
	}
	if certsTask.Key != "/etc/ssl/test-app.key" {
		t.Errorf("Key = %q, want %q", certsTask.Key, "/etc/ssl/test-app.key")
	}
	if certsTask.State != StatePresent {
		t.Errorf("State = %q, want %q", certsTask.State, StatePresent)
	}
}
