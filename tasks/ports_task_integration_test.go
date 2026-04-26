package tasks

import (
	"testing"
)

func TestIntegrationPortsAddAndRemove(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-ports"

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
