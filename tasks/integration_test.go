package tasks

import (
	"omakase/subprocess"
	"os"
	"testing"
)

func dokkuAvailable() bool {
	_, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"version"},
	})
	return err == nil
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func skipIfNoDokkuT(t *testing.T) {
	t.Helper()
	if !dokkuAvailable() {
		t.Skip("skipping integration test: dokku not available")
	}
}

func TestIntegrationAppCreateAndDestroy(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-app"

	// ensure clean state
	destroyApp(appName)

	// create the app
	task := AppTask{App: appName, State: StatePresent}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to create app: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new app creation")
	}

	// creating again should be idempotent
	result = task.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent create failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing app")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// destroy the app
	destroyTask := AppTask{App: appName, State: StateAbsent}
	result = destroyTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to destroy app: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for app destruction")
	}

	// destroying again should be idempotent
	result = destroyTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent destroy failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for nonexistent app")
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationConfigSetAndUnset(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-config"

	// ensure clean state
	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set config
	setTask := ConfigTask{
		App:     appName,
		Restart: false,
		Config:  map[string]string{"TEST_KEY": "test_value"},
		State:   StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set config: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new config")
	}

	// setting same config again should be idempotent
	result = setTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent set failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged config")
	}

	// unset config
	unsetTask := ConfigTask{
		App:     appName,
		Restart: false,
		Config:  map[string]string{"TEST_KEY": ""},
		State:   StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset config: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for config removal")
	}

	// unset again should be idempotent
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent unset failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-unset config")
	}
}

func TestIntegrationBuilderProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-builder"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set builder property
	setTask := BuilderPropertyTask{
		App:      appName,
		Property: "selected",
		Value:    "dockerfile",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set builder property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset builder property
	unsetTask := BuilderPropertyTask{
		App:      appName,
		Property: "selected",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset builder property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationNetworkProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-network"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set network property
	setTask := NetworkPropertyTask{
		App:      appName,
		Property: "bind-all-interfaces",
		Value:    "true",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set network property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset network property
	unsetTask := NetworkPropertyTask{
		App:      appName,
		Property: "bind-all-interfaces",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset network property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationStorageEnsure(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-storage"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	task := StorageEnsureTask{
		App:   appName,
		Chown: "herokuish",
		State: StatePresent,
	}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to ensure storage: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
}

func TestIntegrationStorageMount(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-mount"
	hostDir := "/var/lib/dokku/data/storage/omakase-test-mount"
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

func TestIntegrationGetTasksFullWorkflow(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-workflow"

	// ensure clean state
	destroyApp(appName)
	defer destroyApp(appName)

	data := []byte(`---
- tasks:
    - name: create app
      dokku_app:
        app: ` + appName + `
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	for _, name := range tasks.Keys() {
		task := tasks.Get(name)
		state := task.Execute()
		if state.Error != nil {
			t.Fatalf("task %q failed: %v", name, state.Error)
		}
		if state.State != task.DesiredState() {
			t.Errorf("task %q: expected state %q, got %q", name, task.DesiredState(), state.State)
		}
	}
}
