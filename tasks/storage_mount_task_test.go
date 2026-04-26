package tasks

import (
	"testing"
)

func TestStorageMountTaskInvalidState(t *testing.T) {
	task := StorageMountTask{
		App:          "test-app",
		HostDir:      "/host",
		ContainerDir: "/container",
		State:        "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}
