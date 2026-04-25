package tasks

import (
	"omakase/subprocess"
	"os"
	"strings"
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

func dokkuPluginInstalled(plugin string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"plugin:list"},
	})
	if err != nil {
		return false
	}

	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == plugin {
			return true
		}
	}
	return false
}

func skipIfPluginMissingT(t *testing.T, plugin string) {
	t.Helper()
	if !dokkuPluginInstalled(plugin) {
		t.Skipf("skipping integration test: dokku plugin %q not installed", plugin)
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

func TestIntegrationChecksToggle(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-checks"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// enable checks
	enableTask := ChecksToggleTask{App: appName, State: StatePresent}
	result := enableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to enable checks: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// disable checks
	disableTask := ChecksToggleTask{App: appName, State: StateAbsent}
	result = disableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to disable checks: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationDomainsToggle(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-domains"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// enable domains
	enableTask := DomainsToggleTask{App: appName, State: StatePresent}
	result := enableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to enable domains: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// disable domains
	disableTask := DomainsToggleTask{App: appName, State: StateAbsent}
	result = disableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to disable domains: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationProxyToggle(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-proxy"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// enable proxy
	enableTask := ProxyToggleTask{App: appName, State: StatePresent}
	result := enableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to enable proxy: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// disable proxy
	disableTask := ProxyToggleTask{App: appName, State: StateAbsent}
	result = disableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to disable proxy: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationPortsAddAndRemove(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-ports"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	portMappings := []PortMapping{
		{Scheme: "http", Host: 8080, Container: 5000},
	}

	// add port
	addTask := PortsTask{App: appName, PortMappings: portMappings, State: StatePresent}
	result := addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to add port: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new port")
	}

	// add same port again (idempotent)
	result = addTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent add failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing port")
	}

	// remove port
	removeTask := PortsTask{App: appName, PortMappings: portMappings, State: StateAbsent}
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to remove port: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for port removal")
	}

	// remove again (idempotent)
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent remove failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for nonexistent port")
	}
}

func TestIntegrationConfigMultipleKeys(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-multiconfig"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set 3 keys
	setTask := ConfigTask{
		App:     appName,
		Restart: false,
		Config:  map[string]string{"KEY_A": "val_a", "KEY_B": "val_b", "KEY_C": "val_c"},
		State:   StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set config: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for new config keys")
	}

	// update one key, keep others the same
	updateTask := ConfigTask{
		App:     appName,
		Restart: false,
		Config:  map[string]string{"KEY_A": "val_a", "KEY_B": "val_b_updated", "KEY_C": "val_c"},
		State:   StatePresent,
	}
	result = updateTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to update config: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for partial config update")
	}

	// set same values again (idempotent)
	result = updateTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent update failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged config")
	}

	// unset all keys
	unsetTask := ConfigTask{
		App:     appName,
		Restart: false,
		Config:  map[string]string{"KEY_A": "", "KEY_B": "", "KEY_C": ""},
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
}

func TestIntegrationGitSync(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-gitsync"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	task := GitSyncTask{
		App:    appName,
		Remote: "https://github.com/dokku/smoke-test-app",
		State:  StatePresent,
	}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to sync git: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for git sync")
	}
}

func TestIntegrationGitFromImage(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-fromimage"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	task := GitFromImageTask{
		App:   appName,
		Image: "nginx:latest",
		State: StateDeployed,
	}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to deploy from image: %v", result.Error)
	}
	if result.State != StateDeployed {
		t.Errorf("expected state 'deployed', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for initial deploy")
	}

	// deploy same image again (idempotent)
	result = task.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent deploy failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for same image")
	}
	if result.State != StateDeployed {
		t.Errorf("expected state 'deployed', got '%s'", result.State)
	}
}

func TestIntegrationResourceLimit(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-reslimit"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set resource limits
	setTask := ResourceLimitTask{
		App:       appName,
		Resources: map[string]string{"cpu": "100", "memory": "256"},
		State:     StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set resource limits: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new resource limits")
	}

	// setting same limits again should be idempotent
	result = setTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent set failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged resource limits")
	}

	// clear resource limits
	clearTask := ResourceLimitTask{
		App:   appName,
		State: StateAbsent,
	}
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear resource limits: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for clearing resource limits")
	}

	// clear again should be idempotent
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent clear failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-cleared resource limits")
	}
}

func TestIntegrationResourceLimitProcessType(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-reslimit-pt"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set resource limits for a specific process type
	setTask := ResourceLimitTask{
		App:         appName,
		ProcessType: "web",
		Resources:   map[string]string{"memory": "512"},
		State:       StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set resource limits: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for new resource limits")
	}

	// idempotent
	result = setTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent set failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged resource limits")
	}

	// clear before + set new values
	clearBeforeTask := ResourceLimitTask{
		App:         appName,
		ProcessType: "web",
		Resources:   map[string]string{"cpu": "200"},
		ClearBefore: true,
		State:       StatePresent,
	}
	result = clearBeforeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear_before and set: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for clear_before operation")
	}

	// clear process-type limits
	clearTask := ResourceLimitTask{
		App:         appName,
		ProcessType: "web",
		State:       StateAbsent,
	}
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for clearing resource limits")
	}
}

func TestIntegrationResourceReserve(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-resreserve"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set resource reservations
	setTask := ResourceReserveTask{
		App:       appName,
		Resources: map[string]string{"cpu": "100", "memory": "256"},
		State:     StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set resource reservations: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new resource reservations")
	}

	// setting same reservations again should be idempotent
	result = setTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent set failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged resource reservations")
	}

	// clear resource reservations
	clearTask := ResourceReserveTask{
		App:   appName,
		State: StateAbsent,
	}
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear resource reservations: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for clearing resource reservations")
	}

	// clear again should be idempotent
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent clear failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-cleared resource reservations")
	}
}

func TestIntegrationResourceReserveProcessType(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-resreserve-pt"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set resource reservations for a specific process type
	setTask := ResourceReserveTask{
		App:         appName,
		ProcessType: "web",
		Resources:   map[string]string{"memory": "512"},
		State:       StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set resource reservations: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for new resource reservations")
	}

	// idempotent
	result = setTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent set failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged resource reservations")
	}

	// clear before + set new values
	clearBeforeTask := ResourceReserveTask{
		App:         appName,
		ProcessType: "web",
		Resources:   map[string]string{"cpu": "200"},
		ClearBefore: true,
		State:       StatePresent,
	}
	result = clearBeforeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear_before and set: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for clear_before operation")
	}

	// clear process-type reservations
	clearTask := ResourceReserveTask{
		App:         appName,
		ProcessType: "web",
		State:       StateAbsent,
	}
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for clearing resource reservations")
	}
}

func TestIntegrationMultiTaskWorkflow(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-multi"

	destroyApp(appName)
	defer destroyApp(appName)

	data := []byte(`---
- tasks:
    - name: create app
      dokku_app:
        app: ` + appName + `
    - name: set config
      dokku_config:
        app: ` + appName + `
        restart: false
        config:
          TEST_KEY: test_value
    - name: ensure storage
      dokku_storage_ensure:
        app: ` + appName + `
        chown: herokuish
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	if len(tasks.Keys()) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks.Keys()))
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
		if !state.Changed {
			t.Errorf("task %q: expected changed=true on first run", name)
		}
	}
}

func TestIntegrationServiceCreateAndDestroy(t *testing.T) {
	skipIfNoDokkuT(t)
	skipIfPluginMissingT(t, "redis")

	serviceName := "omakase-test-service"
	serviceType := "redis"

	// ensure clean state
	destroyService(serviceType, serviceName)

	// create the service
	task := ServiceCreateTask{Service: serviceType, Name: serviceName, State: StatePresent}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to create service: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new service creation")
	}

	// creating again should be idempotent
	result = task.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent create failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing service")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// destroy the service
	destroyTask := ServiceCreateTask{Service: serviceType, Name: serviceName, State: StateAbsent}
	result = destroyTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to destroy service: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for service destruction")
	}

	// destroying again should be idempotent
	result = destroyTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent destroy failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for nonexistent service")
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
