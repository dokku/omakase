package commands

import (
	"reflect"
	"testing"

	"github.com/dokku/docket/tasks"
)

func TestApplyCommandMetadata(t *testing.T) {
	c := &ApplyCommand{}
	if c.Name() != "apply" {
		t.Errorf("Name = %q, want \"apply\"", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis must not be empty")
	}
}

func TestApplyCommandHelpDoesNotPanic(t *testing.T) {
	c := &ApplyCommand{}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FlagSet panicked without tasks.yml on disk: %v", r)
		}
	}()
	_ = c.FlagSet()
}

// TestEnvelopeExprContextSkipsNonLoopEnvelopes ensures the per-envelope
// expr context only injects `.item` / `.index` for loop expansions; a
// non-loop envelope evaluates `when:` against the file-level inputs only.
func TestEnvelopeExprContextSkipsNonLoopEnvelopes(t *testing.T) {
	base := map[string]interface{}{"env": "prod"}
	env := &tasks.TaskEnvelope{Name: "x"}
	got := envelopeExprContext(base, env)
	if !reflect.DeepEqual(got, base) {
		t.Errorf("non-loop envelope must return base context unchanged; got %v", got)
	}
}

func TestEnvelopeExprContextInjectsLoopVars(t *testing.T) {
	base := map[string]interface{}{"env": "prod"}
	env := &tasks.TaskEnvelope{Name: "x", IsLoopExpansion: true, LoopItem: "api", LoopIndex: 2}
	got := envelopeExprContext(base, env)
	if got["item"] != "api" {
		t.Errorf("item = %v, want api", got["item"])
	}
	if got["index"] != 2 {
		t.Errorf("index = %v, want 2", got["index"])
	}
	if got["env"] != "prod" {
		t.Errorf("env should be inherited; got %v", got["env"])
	}
	if base["item"] != nil {
		t.Errorf("base must not be mutated by envelopeExprContext")
	}
}

// TestFilterByTagsThroughCommandsLayer covers the apply / plan path's
// reliance on tasks.FilterByTags. Routing the flags through the helper
// in the tasks package keeps the command-side wiring trivial; this test
// pins the behaviour the commands rely on.
func TestFilterByTagsThroughCommandsLayer(t *testing.T) {
	m := tasks.OrderedStringEnvelopeMap{}
	m.Set("api", &tasks.TaskEnvelope{Name: "api", Tags: []string{"api"}})
	m.Set("worker", &tasks.TaskEnvelope{Name: "worker", Tags: []string{"worker"}})
	m.Set("untagged", &tasks.TaskEnvelope{Name: "untagged"})

	if got := tasks.FilterByTags(m, []string{"api"}, nil); !reflect.DeepEqual(got, []string{"api"}) {
		t.Errorf("tags=[api] got %v, want [api]", got)
	}
	if got := tasks.FilterByTags(m, nil, []string{"api"}); !reflect.DeepEqual(got, []string{"worker", "untagged"}) {
		t.Errorf("skip-tags=[api] got %v, want [worker untagged]", got)
	}
	if got := tasks.FilterByTags(m, nil, nil); len(got) != 3 {
		t.Errorf("no flags: got %v, want all 3", got)
	}
}
