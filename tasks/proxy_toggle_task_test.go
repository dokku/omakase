package tasks

import (
	"testing"
)

func TestProxyToggleTaskInvalidState(t *testing.T) {
	task := ProxyToggleTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}
