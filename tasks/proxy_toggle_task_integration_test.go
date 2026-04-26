package tasks

import (
	"testing"
)

func TestIntegrationProxyToggle(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-proxy"

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
