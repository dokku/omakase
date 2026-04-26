package tasks

import (
	"docket/subprocess"
	"strings"
	"testing"
)

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

	networkName := "docket-test-network"

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
