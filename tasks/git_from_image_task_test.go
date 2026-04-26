package tasks

import (
	"strings"
	"testing"
)

func TestGitFromImageTaskInvalidState(t *testing.T) {
	task := GitFromImageTask{App: "test-app", Image: "nginx", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestGitFromImageTaskNonDeployedStates(t *testing.T) {
	tests := []struct {
		name  string
		state State
	}{
		{"present state", StatePresent},
		{"absent state", StateAbsent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := GitFromImageTask{App: "test-app", Image: "nginx", State: tt.state}
			result := task.Execute()
			if result.Error == nil {
				t.Fatal("expected error for non-deployed state")
			}
			if !strings.Contains(result.Error.Error(), "invalid state") {
				t.Errorf("expected 'invalid state' error, got: %v", result.Error)
			}
		})
	}
}
