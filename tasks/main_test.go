package tasks

import (
	"strings"
	"testing"
)

func TestGetTasksEmptyRecipe(t *testing.T) {
	data := []byte("---\n")
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("GetTasks with empty recipe should return an error")
	}

	if !strings.Contains(err.Error(), "no recipe found") {
		t.Errorf("expected 'no recipe found' error, got: %v", err)
	}
}

func TestGetTasksEmptyList(t *testing.T) {
	data := []byte("---\n- tasks: []\n")
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks with empty task list should not error, got: %v", err)
	}

	if len(tasks.Keys()) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks.Keys()))
	}
}

func TestGetTasksValidAppTask(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create test app
      dokku_app:
        app: test-app
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks should not error for valid app task, got: %v", err)
	}

	if len(tasks.Keys()) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks.Keys()))
	}

	task := tasks.Get("create test app")
	if task == nil {
		t.Fatal("task 'create test app' not found")
	}

	if task.DesiredState() != StatePresent {
		t.Errorf("expected state 'present', got '%s'", task.DesiredState())
	}
}

func TestGetTasksInvalidTaskType(t *testing.T) {
	data := []byte(`---
- tasks:
    - dokku_nonexistent:
        app: test-app
`)
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("GetTasks with invalid task type should return an error")
	}

	if !strings.Contains(err.Error(), "not a valid task") {
		t.Errorf("expected 'not a valid task' error, got: %v", err)
	}
}

func TestGetTasksTooManyProperties(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: test
      dokku_app:
        app: test-app
      dokku_config:
        app: test-app
`)
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("GetTasks with too many properties should return an error")
	}

	if !strings.Contains(err.Error(), "too many properties") {
		t.Errorf("expected 'too many properties' error, got: %v", err)
	}
}

func TestGetTasksAutoGeneratesName(t *testing.T) {
	data := []byte(`---
- tasks:
    - dokku_app:
        app: test-app
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks should not error for nameless task, got: %v", err)
	}

	keys := tasks.Keys()
	if len(keys) != 1 {
		t.Fatalf("expected 1 task, got %d", len(keys))
	}

	if !strings.HasPrefix(keys[0], "task #1 ") {
		t.Errorf("expected auto-generated name starting with 'task #1 ', got '%s'", keys[0])
	}
}

func TestGetTasksWithTemplateContext(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create {{ .app_name }}
      dokku_app:
        app: {{ .app_name }}
`)
	context := map[string]interface{}{
		"app_name": "my-app",
	}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks should not error with template context, got: %v", err)
	}

	if len(tasks.Keys()) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks.Keys()))
	}

	task := tasks.Get("create my-app")
	if task == nil {
		t.Fatal("task 'create my-app' not found")
	}
}

func TestRegisteredTasksExist(t *testing.T) {
	expectedTasks := []string{
		"dokku_app",
		"dokku_builder_property",
		"dokku_checks_toggle",
		"dokku_config",
		"dokku_domains_toggle",
		"dokku_git_from_image",
		"dokku_git_sync",
		"dokku_network_property",
		"dokku_ports",
		"dokku_proxy_toggle",
		"dokku_storage_ensure",
		"dokku_storage_mount",
	}

	for _, name := range expectedTasks {
		if _, ok := RegisteredTasks[name]; !ok {
			t.Errorf("expected task %q to be registered", name)
		}
	}
}
