package tasks

import (
	"testing"

	_ "github.com/gliderlabs/sigil/builtin"
)

func TestConfigTaskInvalidState(t *testing.T) {
	task := ConfigTask{
		App:    "test-app",
		Config: map[string]string{"KEY": "VALUE"},
		State:  "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestGetTasksConfigTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: set config
      dokku_config:
        app: test-app
        restart: false
        config:
          KEY1: val1
          KEY2: val2
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("set config")
	if task == nil {
		t.Fatal("task 'set config' not found")
	}

	configTask, ok := task.(*ConfigTask)
	if !ok {
		// tasks may be stored as value types depending on reflection
		ct, ok2 := task.(ConfigTask)
		if !ok2 {
			t.Fatalf("task is not a ConfigTask (type is %T)", task)
		}
		configTask = &ct
	}

	if configTask.App != "test-app" {
		t.Errorf("App = %q, want %q", configTask.App, "test-app")
	}
	// Note: defaults.SetDefaults overrides restart=false with the default tag value "true"
	// because false is the zero value for bool. This documents the actual behavior.
	if !configTask.Restart {
		t.Error("Restart = false, want true (defaults.SetDefaults overrides zero-value bool)")
	}
	if len(configTask.Config) != 2 {
		t.Fatalf("expected 2 config keys, got %d", len(configTask.Config))
	}
	if configTask.Config["KEY1"] != "val1" {
		t.Errorf("Config[KEY1] = %q, want %q", configTask.Config["KEY1"], "val1")
	}
	if configTask.Config["KEY2"] != "val2" {
		t.Errorf("Config[KEY2] = %q, want %q", configTask.Config["KEY2"], "val2")
	}
}
