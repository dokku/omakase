package tasks

import (
	"errors"
	"testing"
)

// TestAllTasksImplementPlan asserts every registered task type satisfies the
// Task interface (which requires Plan()). Without this enforcement, a new
// task that forgets Plan() would only fail when its file is loaded by the
// CLI, instead of at compile time. The Task interface check itself is a
// compile-time guarantee, so this test mostly documents intent and catches
// any future refactor that loosens the interface contract.
func TestAllTasksImplementPlan(t *testing.T) {
	if len(RegisteredTasks) == 0 {
		t.Fatal("no tasks registered; init() side effects didn't run")
	}
	for name, task := range RegisteredTasks {
		var _ Task = task
		// Calling Plan() on a zero-valued task can hit subprocess (which
		// fails in CI without dokku) - we only care that the method exists,
		// which the interface assignment above guarantees. Surface name in
		// errors for any future inspection-style assertions.
		if name == "" {
			t.Errorf("task with empty name: %T", task)
		}
	}
}

// TestDispatchPlanInvalidState verifies that DispatchPlan returns an error
// PlanResult when the requested state has no handler, mirroring DispatchState.
func TestDispatchPlanInvalidState(t *testing.T) {
	result := DispatchPlan(State("nonsense"), map[State]func() PlanResult{
		StatePresent: func() PlanResult { return PlanResult{InSync: true} },
	})
	if result.Error == nil {
		t.Fatal("DispatchPlan with unknown state should set Error")
	}
	if result.Status != PlanStatusError {
		t.Errorf("expected PlanStatusError, got %q", result.Status)
	}
}

// TestDispatchPlanSetsDesiredState verifies DispatchPlan propagates the
// requested state onto the returned PlanResult, matching DispatchState.
func TestDispatchPlanSetsDesiredState(t *testing.T) {
	result := DispatchPlan(StatePresent, map[State]func() PlanResult{
		StatePresent: func() PlanResult { return PlanResult{InSync: true, Status: PlanStatusOK} },
	})
	if result.DesiredState != StatePresent {
		t.Errorf("expected DesiredState=%q, got %q", StatePresent, result.DesiredState)
	}
	if !result.InSync {
		t.Error("expected handler's InSync to propagate")
	}
}

// TestPlanResultError shows the canonical shape of an error PlanResult so
// future helpers that build error results stay consistent.
func TestPlanResultError(t *testing.T) {
	want := errors.New("boom")
	got := PlanResult{Status: PlanStatusError, Error: want}
	if got.Error != want {
		t.Errorf("Error not preserved")
	}
	if got.InSync {
		t.Error("error result must not be InSync")
	}
}

// TestPlanConfigSetItemizesKeys sanity-checks that the config_task helper
// reports the expected per-key mutations when given a synthetic current/desired
// pair. This guards the documented "rich Mutations everywhere" promise for the
// config task family.
func TestPlanConfigSetItemizesKeys(t *testing.T) {
	current := map[string]string{
		"EXISTING": "old",
		"KEEP":     "same",
	}
	desired := map[string]string{
		"EXISTING": "new",
		"KEEP":     "same",
		"NEW":      "value",
	}

	keys := configKeysToSet(current, desired)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys to set, got %d: %v", len(keys), keys)
	}

	got := map[string]bool{}
	for _, k := range keys {
		got[k] = true
	}
	if !got["EXISTING"] {
		t.Error("expected EXISTING in keys to set")
	}
	if !got["NEW"] {
		t.Error("expected NEW in keys to set")
	}
	if got["KEEP"] {
		t.Error("KEEP should not be in keys to set (value matches)")
	}
}

// TestPlanConfigUnsetItemizesKeys sanity-checks that the unset path itemizes
// only keys that currently exist on the server, since unsetting a missing key
// is a no-op.
func TestPlanConfigUnsetItemizesKeys(t *testing.T) {
	current := map[string]string{"PRESENT": "v"}
	desired := map[string]string{"PRESENT": "", "MISSING": ""}

	keys := configKeysToUnset(current, desired)
	if len(keys) != 1 || keys[0] != "PRESENT" {
		t.Errorf("expected only PRESENT in keys to unset, got %v", keys)
	}
}

// TestPlanResultStatusConstants documents the canonical status values used
// across the codebase. If a refactor renames or removes one, this test fails
// loudly so the plan output formatter and JSON consumers can be updated.
func TestPlanResultStatusConstants(t *testing.T) {
	cases := map[PlanStatus]string{
		PlanStatusOK:      "ok",
		PlanStatusModify:  "~",
		PlanStatusCreate:  "+",
		PlanStatusDestroy: "-",
		PlanStatusError:   "error",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("status %q has unexpected string %q", want, string(got))
		}
	}
}
