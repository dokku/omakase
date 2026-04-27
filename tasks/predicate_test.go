package tasks

import (
	"strings"
	"testing"
)

func TestCompilePredicateEmptyReturnsNil(t *testing.T) {
	prog, err := CompilePredicate("")
	if err != nil {
		t.Fatalf("CompilePredicate(empty) err = %v", err)
	}
	if prog != nil {
		t.Fatalf("CompilePredicate(empty) prog = %v, want nil", prog)
	}
}

func TestCompilePredicateSyntaxErrorReports(t *testing.T) {
	_, err := CompilePredicate("env ==")
	if err == nil {
		t.Fatal("expected syntax error, got nil")
	}
	if !strings.Contains(err.Error(), "(") {
		t.Errorf("expected error to include position info, got: %v", err)
	}
}

func TestCompilePredicateCacheReturnsSamePointer(t *testing.T) {
	const src = "env == \"prod\""

	a, err := CompilePredicate(src)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}
	b, err := CompilePredicate(src)
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}
	if a != b {
		t.Fatalf("cache miss: a=%p b=%p", a, b)
	}
}

func TestEvalBoolTruthy(t *testing.T) {
	prog, err := CompilePredicate("env == \"prod\"")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	got, err := EvalBool(prog, map[string]interface{}{"env": "prod"})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if !got {
		t.Errorf("got false, want true")
	}
}

func TestEvalBoolFalsy(t *testing.T) {
	prog, err := CompilePredicate("env == \"prod\"")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	got, err := EvalBool(prog, map[string]interface{}{"env": "staging"})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got {
		t.Errorf("got true, want false")
	}
}

func TestEvalBoolNilProgramIsTrue(t *testing.T) {
	got, err := EvalBool(nil, nil)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if !got {
		t.Errorf("nil program must evaluate as true (no predicate)")
	}
}

func TestEvalListReturnsSlice(t *testing.T) {
	prog, err := CompilePredicate("[\"a\", \"b\", \"c\"]")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	got, err := EvalList(prog, map[string]interface{}{})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
}

func TestEvalListNonListErrors(t *testing.T) {
	prog, err := CompilePredicate("\"not a list\"")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err := EvalList(prog, nil); err == nil {
		t.Fatal("expected non-list error, got nil")
	}
}
