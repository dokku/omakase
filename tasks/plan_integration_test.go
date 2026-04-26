package tasks

import (
	"testing"
)

// TestIntegrationPlanDetectsMissingApp asserts Plan() reports drift for a
// dokku_app task when the target app does not exist, and stays consistent
// with apply: after running apply, a follow-up Plan() returns InSync.
func TestIntegrationPlanDetectsMissingApp(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-plan-detect"
	destroyApp(appName)
	defer destroyApp(appName)

	task := AppTask{App: appName, State: StatePresent}

	plan := task.Plan()
	if plan.Error != nil {
		t.Fatalf("Plan() errored: %v", plan.Error)
	}
	if plan.InSync {
		t.Errorf("expected drift on missing app, got InSync=true")
	}
	if plan.Status != PlanStatusCreate {
		t.Errorf("expected status=%q, got %q", PlanStatusCreate, plan.Status)
	}
	if len(plan.Mutations) == 0 {
		t.Error("expected at least one mutation entry")
	}
}

// TestIntegrationPlanInSyncAfterApply applies a single dokku_app task then
// verifies Plan() reports InSync. This is the round-trip property issue #198
// promises: every apply followed by an immediate plan reports clean.
func TestIntegrationPlanInSyncAfterApply(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-plan-roundtrip"
	destroyApp(appName)
	defer destroyApp(appName)

	task := AppTask{App: appName, State: StatePresent}
	if state := task.Execute(); state.Error != nil {
		t.Fatalf("apply errored: %v", state.Error)
	}

	plan := task.Plan()
	if plan.Error != nil {
		t.Fatalf("plan errored after apply: %v", plan.Error)
	}
	if !plan.InSync {
		t.Errorf("expected InSync after apply, got %+v", plan)
	}
	if plan.Status != PlanStatusOK {
		t.Errorf("expected status=%q, got %q", PlanStatusOK, plan.Status)
	}
}

// TestIntegrationPlanDoesNotMutate is the safety contract for plan: it must
// never create, modify, or destroy server state. We assert this for the
// create case (state=present, app missing).
func TestIntegrationPlanDoesNotMutate(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-plan-no-mutate"
	destroyApp(appName)
	defer destroyApp(appName)

	task := AppTask{App: appName, State: StatePresent}
	_ = task.Plan()

	if appExists(appName) {
		t.Errorf("Plan() unexpectedly created %s", appName)
	}
}

// TestIntegrationPlanConfigItemizes asserts the per-key Mutations contract for
// multi-mutation tasks. The config task is the canonical case.
func TestIntegrationPlanConfigItemizes(t *testing.T) {
	skipIfNoDokkuT(t)

	appName := "docket-test-plan-config"
	destroyApp(appName)
	defer destroyApp(appName)

	if state := (AppTask{App: appName, State: StatePresent}).Execute(); state.Error != nil {
		t.Fatalf("setup apps:create failed: %v", state.Error)
	}

	task := ConfigTask{
		App:     appName,
		Restart: false,
		Config:  map[string]string{"DOCKET_TEST_KEY_A": "1", "DOCKET_TEST_KEY_B": "2"},
		State:   StatePresent,
	}

	plan := task.Plan()
	if plan.Error != nil {
		t.Fatalf("plan errored: %v", plan.Error)
	}
	if plan.InSync {
		t.Fatal("expected drift, got InSync=true")
	}
	if len(plan.Mutations) != 2 {
		t.Errorf("expected 2 mutations (one per key), got %d: %v", len(plan.Mutations), plan.Mutations)
	}
}
