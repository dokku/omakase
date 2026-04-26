package tasks

import (
	"testing"

	_ "github.com/gliderlabs/sigil/builtin"
)

func TestResourceLimitTaskInvalidState(t *testing.T) {
	task := ResourceLimitTask{
		App:       "test-app",
		Resources: map[string]string{"cpu": "100"},
		State:     "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestResourceLimitTaskEmptyResources(t *testing.T) {
	task := ResourceLimitTask{App: "test-app", Resources: map[string]string{}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with empty resources and state=present should return an error")
	}
}

func TestResourceLimitTaskNilResources(t *testing.T) {
	task := ResourceLimitTask{App: "test-app", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with nil resources and state=present should return an error")
	}
}

func TestGetTasksResourceLimitTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: set resource limits
      dokku_resource_limit:
        app: test-app
        process_type: web
        resources:
          cpu: "100"
          memory: "256"
        clear_before: true
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("set resource limits")
	if task == nil {
		t.Fatal("task 'set resource limits' not found")
	}

	rlTask, ok := task.(*ResourceLimitTask)
	if !ok {
		rt, ok2 := task.(ResourceLimitTask)
		if !ok2 {
			t.Fatalf("task is not a ResourceLimitTask (type is %T)", task)
		}
		rlTask = &rt
	}

	if rlTask.App != "test-app" {
		t.Errorf("App = %q, want %q", rlTask.App, "test-app")
	}
	if rlTask.ProcessType != "web" {
		t.Errorf("ProcessType = %q, want %q", rlTask.ProcessType, "web")
	}
	if len(rlTask.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(rlTask.Resources))
	}
	if rlTask.Resources["cpu"] != "100" {
		t.Errorf("Resources[cpu] = %q, want %q", rlTask.Resources["cpu"], "100")
	}
	if rlTask.Resources["memory"] != "256" {
		t.Errorf("Resources[memory] = %q, want %q", rlTask.Resources["memory"], "256")
	}
	// Unlike ConfigTask.Restart (default:"true"), ClearBefore has default:"false" which
	// is the zero value for bool. Since defaults.SetDefaults only overrides zero values,
	// and true is non-zero, setting clear_before: true in YAML is preserved correctly.
	if !rlTask.ClearBefore {
		t.Error("ClearBefore = false, want true (YAML value should be preserved)")
	}
}
