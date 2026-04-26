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

	assertListed := func(t *testing.T, label string, want map[string]bool) {
		t.Helper()
		got, err := getBuildpacks(appName)
		if err != nil {
			t.Fatalf("%s: getBuildpacks failed: %v", label, err)
		}
		if len(got) != len(want) {
			t.Errorf("%s: buildpacks list size = %d, want %d (got=%v)", label, len(got), len(want), got)
		}
		for k := range want {
			if !got[k] {
				t.Errorf("%s: expected buildpack %q to be listed (got=%v)", label, k, got)
			}
		}
		for k := range got {
			if !want[k] {
				t.Errorf("%s: unexpected buildpack %q listed", label, k)
			}
		}
	}

	// initial state - empty
	assertListed(t, "initial", map[string]bool{})

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
	assertListed(t, "after add", map[string]bool{bp: true})

	// add same buildpack again - should be idempotent
	result = addTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second add: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent add")
	}
	assertListed(t, "after idempotent add", map[string]bool{bp: true})

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
	assertListed(t, "after remove", map[string]bool{})

	// remove same buildpack again - should be idempotent
	result = removeTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second remove: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent remove")
	}
	assertListed(t, "after idempotent remove", map[string]bool{})

	// add buildpack again, then clear
	if err := addTask.Execute().Error; err != nil {
		t.Fatalf("failed to re-add buildpack: %v", err)
	}
	assertListed(t, "after re-add", map[string]bool{bp: true})

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
	assertListed(t, "after clear", map[string]bool{})

	// clear again - should be idempotent
	result = clearTask.Execute()
	if result.Error != nil {
		t.Fatalf("failed second clear: %v", result.Error)
	}
	if result.Changed {
		t.Errorf("expected Changed=false on idempotent clear")
	}
	assertListed(t, "after idempotent clear", map[string]bool{})
}
