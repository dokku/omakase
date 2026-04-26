package tasks

import (
	"testing"
)

func TestIntegrationBuildpacks(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-buildpacks"

	destroyApp(appName)
	createApp(appName)
	defer destroyApp(appName)

	bp := "https://github.com/heroku/heroku-buildpack-nodejs.git"

	// add buildpack
	addTask := BuildpacksTask{
		App:        appName,
		Buildpacks: []string{bp},
		State:      StatePresent,
	}
	result := addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to add buildpack: %v", result.Error)
	}
	if !result.Changed {
		t.Errorf("expected Changed=true on first add")
	}
	if result.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", result.State)
	}

	// add same buildpack again - should be idempotent
	result = addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second add: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent add")
	}

	// remove buildpack
	removeTask := BuildpacksTask{
		App:        appName,
		Buildpacks: []string{bp},
		State:      StateAbsent,
	}
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to remove buildpack: %v", result.Error)
	}
	if !result.Changed {
		t.Errorf("expected Changed=true on first remove")
	}
	if result.State != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", result.State)
	}

	// remove same buildpack again - should be idempotent
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second remove: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent remove")
	}

	// add buildpack again, then clear
	if err := addTask.Execute().Error; err != nil {
		t.Fatalf("failed to re-add buildpack: %v", err)
	}

	clearTask := BuildpacksTask{
		App:   appName,
		State: StateAbsent,
	}
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed to clear buildpacks: %v", result.Error)
	}
	if !result.Changed {
		t.Errorf("expected Changed=true on first clear")
	}

	// clear again - should be idempotent
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second clear: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent clear")
	}
}
