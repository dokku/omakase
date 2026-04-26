package tasks

import (
	"strings"
	"testing"
)

func TestDockerOptionsTaskInvalidState(t *testing.T) {
	task := DockerOptionsTask{App: "test-app", Phase: "deploy", Option: "-v /a:/a", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestDockerOptionsTaskMissingApp(t *testing.T) {
	task := DockerOptionsTask{Phase: "deploy", Option: "-v /a:/a", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'app' is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestDockerOptionsTaskInvalidPhase(t *testing.T) {
	for _, phase := range []string{"", "start", "any"} {
		task := DockerOptionsTask{App: "test-app", Phase: phase, Option: "-v /a:/a", State: StatePresent}
		result := task.Execute()
		if result.Error == nil {
			t.Fatalf("Execute with invalid phase %q should return an error", phase)
		}
		if !strings.Contains(result.Error.Error(), "'phase' must be one of") {
			t.Errorf("phase=%q: unexpected error: %v", phase, result.Error)
		}
	}
}

func TestDockerOptionsTaskMissingOption(t *testing.T) {
	task := DockerOptionsTask{App: "test-app", Phase: "deploy", Option: "  ", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without option should return an error")
	}
	if !strings.Contains(result.Error.Error(), "'option' is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestOptionPresent(t *testing.T) {
	tests := []struct {
		existing string
		option   string
		want     bool
	}{
		{"", "-v /a:/a", false},
		{"-v /a:/a", "-v /a:/a", true},
		{"-v /a:/a -p 80:80", "-p 80:80", true},
		{"-v /a:/a -p 80:80", "-v /a:/a", true},
		{"-v /a:/aa", "-v /a:/a", false},      // exact token match, not substring
		{"-v /a:/a", "-v /a:/aa", false},      // exact token match
		{"-p 80:80 -v /a:/a", "-v /a:/a", true},
		{"-p 8080:80", "-p 80:80", false},     // distinct tokens
	}
	for _, tt := range tests {
		got := optionPresent(tt.existing, tt.option)
		if got != tt.want {
			t.Errorf("optionPresent(%q, %q) = %v, want %v", tt.existing, tt.option, got, tt.want)
		}
	}
}

func TestGetTasksDockerOptionsTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: add docker option
      dokku_docker_options:
        app: test-app
        phase: deploy
        option: "-v /var/run/docker.sock:/var/run/docker.sock"
        state: present
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("add docker option")
	if task == nil {
		t.Fatal("task 'add docker option' not found")
	}

	doTask, ok := task.(*DockerOptionsTask)
	if !ok {
		t.Fatalf("task is not a DockerOptionsTask (type is %T)", task)
	}
	if doTask.App != "test-app" {
		t.Errorf("App = %q, want %q", doTask.App, "test-app")
	}
	if doTask.Phase != "deploy" {
		t.Errorf("Phase = %q, want %q", doTask.Phase, "deploy")
	}
	if doTask.Option != "-v /var/run/docker.sock:/var/run/docker.sock" {
		t.Errorf("Option = %q, want %q", doTask.Option, "-v /var/run/docker.sock:/var/run/docker.sock")
	}
	if doTask.State != StatePresent {
		t.Errorf("State = %q, want %q", doTask.State, StatePresent)
	}
}
