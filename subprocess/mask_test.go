package subprocess

import (
	"strings"
	"sync"
	"testing"
)

func TestMaskStringNoGlobalSet(t *testing.T) {
	SetGlobalSensitive(nil)
	defer SetGlobalSensitive(nil)

	if got := MaskString("hello world"); got != "hello world" {
		t.Errorf("MaskString with no global set = %q, want input unchanged", got)
	}
}

func TestMaskStringReplacesAllOccurrences(t *testing.T) {
	SetGlobalSensitive([]string{"secret"})
	defer SetGlobalSensitive(nil)

	got := MaskString("a secret and another secret")
	want := "a *** and another ***"
	if got != want {
		t.Errorf("MaskString = %q, want %q", got, want)
	}
}

func TestMaskStringEmptyEntriesSkipped(t *testing.T) {
	SetGlobalSensitive([]string{"", "tok"})
	defer SetGlobalSensitive(nil)

	got := MaskString("xtoky")
	if got != "x***y" {
		t.Errorf("MaskString = %q, want %q", got, "x***y")
	}
	// Verify the empty entry didn't cause every character to mask.
	if strings.Contains(MaskString("abc"), "***") && !strings.Contains("abc", "tok") {
		t.Errorf("empty entry caused unintended masking")
	}
}

func TestMaskStringLongerBeforeShorter(t *testing.T) {
	// "ab" is a substring of "abcdef"; the longer one must be masked first
	// so we don't see "***cdef" instead of a single "***".
	SetGlobalSensitive([]string{"ab", "abcdef"})
	defer SetGlobalSensitive(nil)

	got := MaskString("xabcdefy")
	if got != "x***y" {
		t.Errorf("MaskString = %q, want %q (longer match first)", got, "x***y")
	}
}

func TestSetGlobalSensitiveDeduplicates(t *testing.T) {
	SetGlobalSensitive([]string{"a", "a", "b"})
	defer SetGlobalSensitive(nil)

	values := GlobalSensitive()
	count := 0
	for _, v := range values {
		if v == "a" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("duplicates not removed: got %d copies of 'a' in %v", count, values)
	}
}

func TestSetGlobalSensitiveClear(t *testing.T) {
	SetGlobalSensitive([]string{"x"})
	SetGlobalSensitive(nil)
	if got := MaskString("xyz"); got != "xyz" {
		t.Errorf("MaskString after clear = %q, want %q", got, "xyz")
	}
}

func TestMaskStringConcurrent(t *testing.T) {
	SetGlobalSensitive([]string{"secret"})
	defer SetGlobalSensitive(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			SetGlobalSensitive([]string{"secret", "token"})
		}()
		go func() {
			defer wg.Done()
			_ = MaskString("a secret token here")
		}()
	}
	wg.Wait()
}

func TestGlobalSensitiveReturnsCopy(t *testing.T) {
	SetGlobalSensitive([]string{"a"})
	defer SetGlobalSensitive(nil)

	values := GlobalSensitive()
	values[0] = "mutated"
	again := GlobalSensitive()
	if again[0] != "a" {
		t.Errorf("GlobalSensitive returned shared slice; mutation leaked: %v", again)
	}
}
