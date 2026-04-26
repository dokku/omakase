package tasks

import (
	"testing"
)

func TestDomainsToggleTaskInvalidState(t *testing.T) {
	task := DomainsToggleTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}
