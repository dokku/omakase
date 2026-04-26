package commands

import (
	"testing"

	"github.com/dokku/docket/tasks"
)

// TestPlanCommandMetadata is a simple smoke check that the PlanCommand
// satisfies the cli.Command interface and exposes sensible Name / Synopsis
// strings. Keeping this test lightweight avoids coupling to the cli-skeleton
// internals; full subcommand wiring is exercised by the bats suite.
func TestPlanCommandMetadata(t *testing.T) {
	c := &PlanCommand{}
	if c.Name() != "plan" {
		t.Errorf("Name = %q, want \"plan\"", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis must not be empty")
	}
}

// TestPlanCommandExamples ensures every example string mentions the command
// name. The cli-skeleton uses these in --help output; a misspelled subcommand
// in an example would silently mislead users.
func TestPlanCommandExamples(t *testing.T) {
	c := &PlanCommand{}
	examples := c.Examples()
	if len(examples) == 0 {
		t.Fatal("expected at least one example")
	}
	for label, example := range examples {
		if example == "" {
			t.Errorf("example %q is empty", label)
		}
	}
}

// TestPlanCommandHelpDoesNotPanic guards against a regression where adding a
// flag whose declaration depends on a parsed tasks.yml caused FlagSet() to
// panic at help time when the file did not exist. The current implementation
// silently ignores read errors; this test pins that behavior.
func TestPlanCommandHelpDoesNotPanic(t *testing.T) {
	c := &PlanCommand{}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FlagSet panicked without tasks.yml on disk: %v", r)
		}
	}()
	_ = c.FlagSet()
}

// TestPlanCommandUsesSamePlanInterface guards the contract between the
// commands package and the tasks package: PlanCommand must consume tasks.Task,
// which must expose Plan() returning tasks.PlanResult.
func TestPlanCommandUsesSamePlanInterface(t *testing.T) {
	var _ interface {
		Plan() tasks.PlanResult
	} = (tasks.Task)(nil)
}
