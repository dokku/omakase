package tasks

import (
	"fmt"
	"omakase/subprocess"
	"os"
	"strconv"
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

func dockerLinkSupported() bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"version", "--format", "{{.Server.Version}}"},
	})
	if err != nil {
		return false
	}

	version := strings.TrimSpace(result.StdoutContents())
	parts := strings.SplitN(version, ".", 2)
	if len(parts) == 0 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	// Docker < 29 supports --link natively
	if major < 29 {
		return true
	}

	// Docker >= 29 requires DOCKER_KEEP_DEPRECATED_LEGACY_LINKS_ENV_VARS=1
	// on the daemon. Test by creating two containers with --link and checking
	// if the link env vars are present.
	subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"rm", "-f", "omakase-link-test-target", "omakase-link-test-client"},
	})

	_, err = subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"run", "-d", "--name", "omakase-link-test-target", "alpine", "sleep", "30"},
	})
	if err != nil {
		return false
	}
	defer subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"rm", "-f", "omakase-link-test-target", "omakase-link-test-client"},
	})

	result, err = subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"run", "--rm", "--name", "omakase-link-test-client", "--link", "omakase-link-test-target:target", "alpine", "env"},
	})
	if err != nil {
		return false
	}

	return strings.Contains(result.StdoutContents(), "TARGET_NAME=")
}

func skipIfDockerLinkUnsupportedT(t *testing.T) {
	t.Helper()
	if !dockerLinkSupported() {
		t.Skip("skipping integration test: docker does not support legacy container links")
	}
}

// getCurrentContainerIDs reads the container IDs from dokku's internal
// CONTAINER files (e.g., /home/dokku/APP/CONTAINER.web.1) which are the
// authoritative source for the current deployment's containers.
func getCurrentContainerIDs(appName, processType string) ([]string, error) {
	scale, err := getPsScale(appName)
	if err != nil {
		return nil, err
	}
	count, ok := scale[processType]
	if !ok || count == 0 {
		return nil, nil
	}
	var ids []string
	for i := 1; i <= count; i++ {
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "cat",
			Args:    []string{fmt.Sprintf("/home/dokku/%s/CONTAINER.%s.%d", appName, processType, i)},
		})
		if err != nil {
			continue
		}
		id := strings.TrimSpace(result.StdoutContents())
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
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

func dockerNetworkExists(name string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"network", "inspect", name, "--format", "{{.Name}}"},
	})
	if err != nil {
		return false
	}
	return strings.TrimSpace(result.StdoutContents()) == name
}

func TestIntegrationNetworkCreateAndDestroy(t *testing.T) {
	skipIfNoDokkuT(t)

	networkName := "omakase-test-network"

	// ensure clean state
	destroyNetwork(networkName)

	// verify network does not exist via docker cli
	if dockerNetworkExists(networkName) {
		t.Fatal("expected network to not exist before creation")
	}

	// create the network
	task := NetworkTask{Name: networkName, State: StatePresent}
	result := task.Execute()
	if result.Error != nil {
		t.Fatalf("failed to create network: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new network creation")
	}

	// verify network exists via docker cli
	if !dockerNetworkExists(networkName) {
		t.Fatal("expected network to exist after creation")
	}

	// verify network driver via docker cli
	inspectResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"network", "inspect", networkName, "--format", "{{.Driver}}"},
	})
	if err != nil {
		t.Fatalf("failed to inspect network driver: %v", err)
	}
	driver := strings.TrimSpace(inspectResult.StdoutContents())
	if driver != "bridge" {
		t.Errorf("expected network driver 'bridge', got '%s'", driver)
	}

	// creating again should be idempotent
	result = task.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent create failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing network")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// destroy the network
	destroyTask := NetworkTask{Name: networkName, State: StateAbsent}
	result = destroyTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to destroy network: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for network destruction")
	}

	// verify network does not exist via docker cli after destroy
	if dockerNetworkExists(networkName) {
		t.Fatal("expected network to not exist after destruction")
	}

	// destroying again should be idempotent
	result = destroyTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent destroy failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for nonexistent network")
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

// getReportedDomains queries dokku domains:report to get the current domain list for an app
func getReportedDomains(appName string) []string {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"domains:report", appName, "--domains-app-vhosts"},
	})
	if err != nil {
		return nil
	}

	return strings.Fields(result.StdoutContents())
}

func TestIntegrationDomainsAddAndRemove(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-domains-task"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// add domains
	addTask := DomainsTask{
		App:     appName,
		Domains: []string{"example.com", "www.example.com"},
		State:   StatePresent,
	}
	result := addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to add domains: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new domains")
	}

	// verify domains via domains:report
	domains := getReportedDomains(appName)
	domainSet := map[string]bool{}
	for _, d := range domains {
		domainSet[d] = true
	}
	if !domainSet["example.com"] {
		t.Error("expected 'example.com' in domains:report output after add")
	}
	if !domainSet["www.example.com"] {
		t.Error("expected 'www.example.com' in domains:report output after add")
	}

	// add same domains again (idempotent)
	result = addTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent add failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing domains")
	}

	// remove one domain
	removeTask := DomainsTask{
		App:     appName,
		Domains: []string{"www.example.com"},
		State:   StateAbsent,
	}
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to remove domain: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for domain removal")
	}

	// verify domains via domains:report after removal
	domains = getReportedDomains(appName)
	domainSet = map[string]bool{}
	for _, d := range domains {
		domainSet[d] = true
	}
	if !domainSet["example.com"] {
		t.Error("expected 'example.com' to still be present after removing www.example.com")
	}
	if domainSet["www.example.com"] {
		t.Error("expected 'www.example.com' to be absent after removal")
	}

	// remove same domain again (idempotent)
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent remove failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-removed domain")
	}

	// set domains (replaces all)
	setTask := DomainsTask{
		App:     appName,
		Domains: []string{"new.example.com"},
		State:   StateSet,
	}
	result = setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set domains: %v", result.Error)
	}
	if result.State != StateSet {
		t.Errorf("expected state 'set', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for set domains")
	}

	// verify domains via domains:report after set
	domains = getReportedDomains(appName)
	if len(domains) != 1 {
		t.Fatalf("expected exactly 1 domain after set, got %d: %v", len(domains), domains)
	}
	if domains[0] != "new.example.com" {
		t.Errorf("expected domain 'new.example.com' after set, got '%s'", domains[0])
	}

	// clear all domains
	clearTask := DomainsTask{
		App:   appName,
		State: StateClear,
	}
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear domains: %v", result.Error)
	}
	if result.State != StateClear {
		t.Errorf("expected state 'clear', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for clear domains")
	}

	// verify no domains via domains:report after clear
	domains = getReportedDomains(appName)
	if len(domains) != 0 {
		t.Errorf("expected 0 domains after clear, got %d: %v", len(domains), domains)
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
		Image: "dokku/smoke-test-app:dockerfile",
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

func TestIntegrationPsScale(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-psscale"

	// ensure clean state
	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// deploy the smoke test app so we have running containers to scale
	deployTask := GitFromImageTask{
		App:   appName,
		Image: "dokku/smoke-test-app:dockerfile",
		State: StateDeployed,
	}
	deployResult := deployTask.Execute()
	if deployResult.Error != nil {
		t.Fatalf("failed to deploy app: %v", deployResult.Error)
	}

	// verify initial web container count is 1 via docker ps
	initialContainers, err := getCurrentContainerIDs(appName, "web")
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}
	if len(initialContainers) != 1 {
		t.Fatalf("expected 1 initial web container, got %d", len(initialContainers))
	}

	// verify the initial container is running via docker inspect
	inspectResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"inspect", "--format", "{{.State.Running}}", initialContainers[0]},
	})
	if err != nil {
		t.Fatalf("failed to inspect initial container: %v", err)
	}
	if strings.TrimSpace(inspectResult.StdoutContents()) != "true" {
		t.Errorf("expected initial container to be running")
	}

	// scale web to 2
	scaleTask := PsScaleTask{
		App:   appName,
		Scale: map[string]int{"web": 2},
		State: StatePresent,
	}
	result := scaleTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to scale app: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for scaling up")
	}

	// clean up old containers and verify 2 web containers via docker ps
	scaledContainers, err := getCurrentContainerIDs(appName, "web")
	if err != nil {
		t.Fatalf("failed to list containers after scale: %v", err)
	}
	if len(scaledContainers) != 2 {
		t.Fatalf("expected 2 web containers after scaling, got %d", len(scaledContainers))
	}

	// verify each container is running via docker inspect
	for _, containerID := range scaledContainers {
		inspectResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "docker",
			Args:    []string{"inspect", "--format", "{{.State.Running}}", containerID},
		})
		if err != nil {
			t.Fatalf("failed to inspect container %s: %v", containerID, err)
		}
		if strings.TrimSpace(inspectResult.StdoutContents()) != "true" {
			t.Errorf("expected container %s to be running", containerID)
		}
	}

	// scaling again should be idempotent
	result = scaleTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent scale failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged scale")
	}

	// scale back to 1
	scaleDownTask := PsScaleTask{
		App:   appName,
		Scale: map[string]int{"web": 1},
		State: StatePresent,
	}
	result = scaleDownTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to scale down: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for scaling down")
	}

	// clean up old containers and verify 1 web container after scale down
	finalContainers, err := getCurrentContainerIDs(appName, "web")
	if err != nil {
		t.Fatalf("failed to list containers after scale down: %v", err)
	}
	if len(finalContainers) != 1 {
		t.Fatalf("expected 1 web container after scale down, got %d", len(finalContainers))
	}

	// verify the final container is running via docker inspect
	inspectResult, err = subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"inspect", "--format", "{{.State.Running}}", finalContainers[0]},
	})
	if err != nil {
		t.Fatalf("failed to inspect final container: %v", err)
	}
	if strings.TrimSpace(inspectResult.StdoutContents()) != "true" {
		t.Errorf("expected final container to be running")
	}
}

func TestIntegrationPsScaleSkipDeploy(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "omakase-test-psscale-sd"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// scale with skip_deploy on an undeployed app
	scaleTask := PsScaleTask{
		App:        appName,
		Scale:      map[string]int{"web": 2, "worker": 1},
		SkipDeploy: true,
		State:      StatePresent,
	}
	result := scaleTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to scale with skip_deploy: %v", result.Error)
	}
	if !result.Changed {
		t.Error("expected changed=true for initial scale")
	}

	// verify the scale was set correctly
	scale, err := getPsScale(appName)
	if err != nil {
		t.Fatalf("failed to get ps scale: %v", err)
	}
	if scale["web"] != 2 {
		t.Errorf("expected web=2, got web=%d", scale["web"])
	}
	if scale["worker"] != 1 {
		t.Errorf("expected worker=1, got worker=%d", scale["worker"])
	}

	// idempotent
	result = scaleTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent scale failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for unchanged scale")
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

func TestIntegrationServiceLinkAndUnlink(t *testing.T) {
	skipIfNoDokkuT(t)
	skipIfPluginMissingT(t, "redis")
	skipIfDockerLinkUnsupportedT(t)

	appName := "omakase-test-link-app"
	serviceName := "omakase-test-link-svc"
	serviceType := "redis"

	// ensure clean state
	destroyApp(appName)
	destroyService(serviceType, serviceName)

	// create prerequisites
	createApp(appName)
	defer destroyApp(appName)

	createTask := ServiceCreateTask{Service: serviceType, Name: serviceName, State: StatePresent}
	createResult := createTask.Execute()
	if createResult.Error != nil {
		t.Fatalf("failed to create service: %v", createResult.Error)
	}
	defer func() {
		// unlink before destroying service
		unlinkTask := ServiceLinkTask{App: appName, Service: serviceType, Name: serviceName, State: StateAbsent}
		unlinkTask.Execute()
		destroyService(serviceType, serviceName)
	}()

	// verify service container is running via docker inspect
	containerName := fmt.Sprintf("dokku.%s.%s", serviceType, serviceName)
	inspectResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"inspect", "--format", "{{.State.Running}}", containerName},
	})
	if err != nil {
		t.Fatalf("failed to inspect service container: %v", err)
	}
	if strings.TrimSpace(inspectResult.StdoutContents()) != "true" {
		t.Errorf("expected service container %q to be running", containerName)
	}

	// link service to app
	linkTask := ServiceLinkTask{App: appName, Service: serviceType, Name: serviceName, State: StatePresent}
	result := linkTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to link service: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for new service link")
	}

	// verify REDIS_URL config var was set by the link
	configResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"config:get", appName, "REDIS_URL"},
	})
	if err != nil {
		t.Fatalf("failed to get REDIS_URL after link: %v", err)
	}
	redisURL := strings.TrimSpace(configResult.StdoutContents())
	if redisURL == "" {
		t.Error("expected REDIS_URL to be set after linking service")
	}
	if !strings.HasPrefix(redisURL, "redis://") {
		t.Errorf("expected REDIS_URL to start with 'redis://', got %q", redisURL)
	}

	// verify the service container exposes the expected network alias via docker inspect
	aliasResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"inspect", "--format", "{{.Config.Hostname}}", containerName},
	})
	if err != nil {
		t.Fatalf("failed to inspect service container hostname: %v", err)
	}
	if strings.TrimSpace(aliasResult.StdoutContents()) == "" {
		t.Error("expected service container to have a hostname set")
	}

	// deploy the smoke test app so we can verify the link inside a running container
	deployTask := GitFromImageTask{
		App:   appName,
		Image: "dokku/smoke-test-app:dockerfile",
		State: StateDeployed,
	}
	deployResult := deployTask.Execute()
	if deployResult.Error != nil {
		t.Fatalf("failed to deploy app: %v", deployResult.Error)
	}

	// find the running app container
	appContainerResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"ps", "--filter", fmt.Sprintf("label=com.dokku.app-name=%s", appName), "--filter", "label=com.dokku.process-type=web", "--format", "{{.ID}}"},
	})
	if err != nil {
		t.Fatalf("failed to find app container: %v", err)
	}
	appContainerID := strings.TrimSpace(appContainerResult.StdoutContents())
	if appContainerID == "" {
		t.Fatal("expected at least one running app container after deploy")
	}
	// take the first container if multiple lines
	appContainerIDs := strings.Split(appContainerID, "\n")
	appContainerID = appContainerIDs[0]

	// verify the app container is running via docker inspect
	appInspectResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"inspect", "--format", "{{.State.Running}}", appContainerID},
	})
	if err != nil {
		t.Fatalf("failed to inspect app container: %v", err)
	}
	if strings.TrimSpace(appInspectResult.StdoutContents()) != "true" {
		t.Error("expected app container to be running")
	}

	// verify REDIS_URL is present inside the running container via docker exec
	execResult, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "docker",
		Args:    []string{"exec", appContainerID, "env"},
	})
	if err != nil {
		t.Fatalf("failed to exec env in app container: %v", err)
	}
	envOutput := execResult.StdoutContents()
	if !strings.Contains(envOutput, "REDIS_URL=redis://") {
		t.Error("expected REDIS_URL=redis://... to be present in app container environment")
	}

	// linking again should be idempotent
	result = linkTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent link failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for existing service link")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unlink service from app
	unlinkTask := ServiceLinkTask{App: appName, Service: serviceType, Name: serviceName, State: StateAbsent}
	result = unlinkTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unlink service: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for service unlink")
	}

	// verify REDIS_URL config var was removed by the unlink
	configResult, err = subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"config:get", appName, "REDIS_URL"},
	})
	if err == nil && strings.TrimSpace(configResult.StdoutContents()) != "" {
		t.Error("expected REDIS_URL to be unset after unlinking service")
	}

	// unlinking again should be idempotent
	result = unlinkTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent unlink failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-unlinked service")
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}

func TestIntegrationHttpAuth(t *testing.T) {
	skipIfNoDokkuT(t)
	skipIfPluginMissingT(t, "http-auth")

	appName := "omakase-test-http-auth"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// enable http auth
	enableTask := HttpAuthTask{
		App:      appName,
		Username: "testuser",
		Password: "testpass",
		State:    StatePresent,
	}
	result := enableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to enable http auth: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for enabling http auth")
	}

	// verify auth is enabled via http-auth:report
	if !httpAuthEnabled(appName) {
		t.Error("expected http auth to be enabled after enable")
	}

	// enabling again should be idempotent
	result = enableTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent enable failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-enabled http auth")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// disable http auth
	disableTask := HttpAuthTask{
		App:   appName,
		State: StateAbsent,
	}
	result = disableTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to disable http auth: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
	if !result.Changed {
		t.Error("expected changed=true for disabling http auth")
	}

	// verify auth is disabled via http-auth:report
	if httpAuthEnabled(appName) {
		t.Error("expected http auth to be disabled after disable")
	}

	// disabling again should be idempotent
	result = disableTask.Execute()
	if result.Error != nil {
		t.Fatalf("idempotent disable failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("expected changed=false for already-disabled http auth")
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
