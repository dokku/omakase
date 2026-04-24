package tasks

import (
	"testing"
)

type mockTask struct {
	name  string
	state State
}

func (m mockTask) DesiredState() State              { return m.state }
func (m mockTask) Doc() string                      { return "" }
func (m mockTask) Examples() ([]Doc, error)         { return nil, nil }
func (m mockTask) Execute() TaskOutputState         { return TaskOutputState{State: m.state} }

func TestOrderedStringTaskMapSetAndGet(t *testing.T) {
	m := OrderedStringTaskMap{}

	task := mockTask{name: "test", state: StatePresent}
	m.Set("key1", task)

	got := m.Get("key1")
	if got == nil {
		t.Fatal("Get returned nil for existing key")
	}

	if got.DesiredState() != StatePresent {
		t.Errorf("expected state 'present', got '%s'", got.DesiredState())
	}
}

func TestOrderedStringTaskMapGetMissing(t *testing.T) {
	m := OrderedStringTaskMap{}

	got := m.Get("nonexistent")
	if got != nil {
		t.Error("Get returned non-nil for nonexistent key on empty map")
	}

	m.Set("key1", mockTask{name: "test", state: StatePresent})
	got = m.Get("nonexistent")
	if got != nil {
		t.Error("Get returned non-nil for nonexistent key")
	}
}

func TestOrderedStringTaskMapKeysPreservesOrder(t *testing.T) {
	m := OrderedStringTaskMap{}

	m.Set("third", mockTask{name: "3", state: StatePresent})
	m.Set("first", mockTask{name: "1", state: StatePresent})
	m.Set("second", mockTask{name: "2", state: StatePresent})

	keys := m.Keys()
	expected := []string{"third", "first", "second"}

	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("key[%d] = %q, want %q", i, key, expected[i])
		}
	}
}

func TestOrderedStringTaskMapEmptyKeys(t *testing.T) {
	m := OrderedStringTaskMap{}

	keys := m.Keys()
	if keys != nil && len(keys) != 0 {
		t.Errorf("expected nil or empty keys, got %v", keys)
	}
}
