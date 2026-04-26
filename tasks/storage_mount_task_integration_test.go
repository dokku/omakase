package tasks

import (
	"docket/subprocess"
	"testing"
)

func TestIntegrationStorageMount(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-mount"
	hostDir := "/var/lib/dokku/data/storage/docket-test-mount"
	containerDir := "/app/storage"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// ensure storage directory exists
	subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "mkdir",
		Args:    []string{"-p", hostDir},
	})

	// mount storage
	mountTask := StorageMountTask{
		App:          appName,
		HostDir:      hostDir,
		ContainerDir: containerDir,
		State:        StatePresent,
	}
	result := mountTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to mount storage: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new mount")
	}

	// mount again should be idempotent
	result = mountTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent mount failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing mount")
	}

	// unmount storage
	unmountTask := StorageMountTask{
		App:          appName,
		HostDir:      hostDir,
		ContainerDir: containerDir,
		State:        StateAbsent,
	}
	result = unmountTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unmount storage: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for unmount")
	}

	// unmount again should be idempotent
	result = unmountTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent unmount failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for nonexistent mount")
	}
}
