package tasks

import (
	"testing"
)

func TestIntegrationHttpAuth(t *testing.T) {
	skipIfNoDokkuT(t)
	skipIfPluginMissingT(t, "http-auth")

	appName := "docket-test-http-auth"

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
