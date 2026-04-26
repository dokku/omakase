package tasks

import (
	"testing"

	_ "github.com/gliderlabs/sigil/builtin"
)

func TestNetworkTaskInvalidState(t *testing.T) {
	task := NetworkTask{Name: "test-network", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestGetTasksNetworkTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create test network
      dokku_network:
        name: test-network
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("create test network")
	if task == nil {
		t.Fatal("task 'create test network' not found")
	}

	netTask, ok := task.(*NetworkTask)
	if !ok {
		nt, ok2 := task.(NetworkTask)
		if !ok2 {
			t.Fatalf("task is not a NetworkTask (type is %T)", task)
		}
		netTask = &nt
	}

	if netTask.Name != "test-network" {
		t.Errorf("Name = %q, want %q", netTask.Name, "test-network")
	}
	if netTask.State != StatePresent {
		t.Errorf("expected default state 'present', got %q", netTask.State)
	}
}
