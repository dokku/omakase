package tasks

import (
	"testing"
)

func TestIntegrationServiceCreateAndDestroy(t *testing.T) {
	skipIfNoDokkuT(t)
	skipIfPluginMissingT(t, "redis")

	serviceName := "docket-test-service"
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
