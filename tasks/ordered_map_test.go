package tasks

import (
	"fmt"
	"testing"
)

type mockTask struct {
	name  string
	state State
}

func (m mockTask) Doc() string              { return "" }
func (m mockTask) Examples() ([]Doc, error) { return nil, nil }
func (m mockTask) Execute() TaskOutputState { return TaskOutputState{State: m.state} }

func TestOrderedStringTaskMapSetAndGet(t *testing.T) {
	m := OrderedStringTaskMap{}

	task := mockTask{name: "test", state: StatePresent}
	m.Set("key1", task)

	got := m.Get("key1")
	if got == nil {
		t.Fatal("Get returned nil for existing key")
	}

	if got.Execute().State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", got.Execute().State)
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

func TestOrderedStringTaskMapOverwriteValue(t *testing.T) {
	m := OrderedStringTaskMap{}

	task1 := mockTask{name: "first", state: StatePresent}
	task2 := mockTask{name: "second", state: StateAbsent}

	m.Set("key", task1)
	m.Set("key", task2)

	// Get should return the latest value
	got := m.Get("key")
	if got == nil {
		t.Fatal("Get returned nil for overwritten key")
	}
	if got.Execute().State != StateAbsent {
		t.Errorf("expected overwritten state 'absent', got '%s'", got.Execute().State)
	}

	// Keys will contain the key twice (current implementation appends without dedup)
	keys := m.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys (duplicate), got %d", len(keys))
	}
}

func TestOrderedStringTaskMapLargeSet(t *testing.T) {
	m := OrderedStringTaskMap{}

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("task-%d", i)
		m.Set(name, mockTask{name: name, state: StatePresent})
	}

	keys := m.Keys()
	if len(keys) != 10 {
		t.Fatalf("expected 10 keys, got %d", len(keys))
	}

	for i, key := range keys {
		expected := fmt.Sprintf("task-%d", i)
		if key != expected {
			t.Errorf("key[%d] = %q, want %q", i, key, expected)
		}
	}
}
