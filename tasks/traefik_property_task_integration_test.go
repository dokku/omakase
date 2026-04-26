package tasks

import (
	"testing"
)

func TestIntegrationTraefikProperty(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-traefik"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	// set traefik property
	setTask := TraefikPropertyTask{
		App:      appName,
		Property: "letsencrypt-email",
		Value:    "admin@example.com",
		State:    StatePresent,
	}
	result := setTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to set traefik property: %v", result.Error)
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// unset traefik property
	unsetTask := TraefikPropertyTask{
		App:      appName,
		Property: "letsencrypt-email",
		State:    StateAbsent,
	}
	result = unsetTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to unset traefik property: %v", result.Error)
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}
}
