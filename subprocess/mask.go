package subprocess

import (
	"sort"
	"strings"
	"sync"
)

// maskPlaceholder is what every sensitive value is replaced with in user-facing
// output. Length is intentionally fixed at three asterisks so masked output
// reveals nothing about the original value (no length, prefix, or suffix).
const maskPlaceholder = "***"

var (
	sensitiveMu     sync.RWMutex
	sensitiveValues []string
)

// SetGlobalSensitive registers the set of literal string values that must be
// masked anywhere they appear in user-facing output. Pass nil or an empty
// slice to clear the registry. Empty entries in values are dropped (matching
// every empty substring would otherwise mask everything).
//
// Callers (typically commands/apply.go and commands/plan.go) collect this set
// from input values declared `sensitive: true` and from task struct fields
// tagged `sensitive:"true"` before any subprocess runs, then defer a clear.
func SetGlobalSensitive(values []string) {
	cleaned := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		cleaned = append(cleaned, v)
	}
	// Sort by length descending so longer matches are masked before any
	// shorter substring of them would be.
	sort.SliceStable(cleaned, func(i, j int) bool {
		return len(cleaned[i]) > len(cleaned[j])
	})

	sensitiveMu.Lock()
	defer sensitiveMu.Unlock()
	if len(cleaned) == 0 {
		sensitiveValues = nil
		return
	}
	sensitiveValues = cleaned
}

// GlobalSensitive returns a snapshot of the current sensitive value set.
func GlobalSensitive() []string {
	sensitiveMu.RLock()
	defer sensitiveMu.RUnlock()
	if len(sensitiveValues) == 0 {
		return nil
	}
	out := make([]string, len(sensitiveValues))
	copy(out, sensitiveValues)
	return out
}

// MaskString replaces every occurrence of any registered sensitive value in s
// with `***`. Returns s unchanged when the registry is empty.
func MaskString(s string) string {
	if s == "" {
		return s
	}
	sensitiveMu.RLock()
	values := sensitiveValues
	sensitiveMu.RUnlock()
	if len(values) == 0 {
		return s
	}
	for _, v := range values {
		if v == "" {
			continue
		}
		s = strings.ReplaceAll(s, v, maskPlaceholder)
	}
	return s
}
