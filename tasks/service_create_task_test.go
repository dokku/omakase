package tasks

import (
	"testing"

	_ "github.com/gliderlabs/sigil/builtin"
)

func TestServiceCreateTaskInvalidState(t *testing.T) {
	task := ServiceCreateTask{Service: "redis", Name: "test-service", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestGetTasksServiceCreateTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create redis service
      dokku_service_create:
        service: redis
        name: my-redis
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("create redis service")
	if task == nil {
		t.Fatal("task 'create redis service' not found")
	}

	scTask, ok := task.(*ServiceCreateTask)
	if !ok {
		st, ok2 := task.(ServiceCreateTask)
		if !ok2 {
			t.Fatalf("task is not a ServiceCreateTask (type is %T)", task)
		}
		scTask = &st
	}

	if scTask.Service != "redis" {
		t.Errorf("Service = %q, want %q", scTask.Service, "redis")
	}
	if scTask.Name != "my-redis" {
		t.Errorf("Name = %q, want %q", scTask.Name, "my-redis")
	}
	if scTask.State != StatePresent {
		t.Errorf("expected default state 'present', got %q", scTask.State)
	}
}

func TestGetTasksServiceCreateWithTemplateContext(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create {{ .service_type }} service
      dokku_service_create:
        service: {{ .service_type }}
        name: {{ .service_name }}
`)
	context := map[string]interface{}{
		"service_type": "postgres",
		"service_name": "my-db",
	}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("create postgres service")
	if task == nil {
		t.Fatal("task 'create postgres service' not found")
	}

	scTask, ok := task.(*ServiceCreateTask)
	if !ok {
		st, ok2 := task.(ServiceCreateTask)
		if !ok2 {
			t.Fatalf("task is not a ServiceCreateTask (type is %T)", task)
		}
		scTask = &st
	}

	if scTask.Service != "postgres" {
		t.Errorf("Service = %q, want %q", scTask.Service, "postgres")
	}
	if scTask.Name != "my-db" {
		t.Errorf("Name = %q, want %q", scTask.Name, "my-db")
	}
}
