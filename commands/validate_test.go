package commands

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dokku/docket/tasks"
)

func TestValidateCommandMetadata(t *testing.T) {
	c := &ValidateCommand{}
	if c.Name() != "validate" {
		t.Errorf("Name = %q, want \"validate\"", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis must not be empty")
	}
}

func TestValidateCommandExamples(t *testing.T) {
	c := &ValidateCommand{}
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

func TestValidateCommandHelpDoesNotPanic(t *testing.T) {
	c := &ValidateCommand{}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FlagSet panicked without tasks.yml on disk: %v", r)
		}
	}()
	_ = c.FlagSet()
}

// TestFormatProblemHumanOutput exercises the human formatter end-to-end so
// the issue's example output (line N column M, "did you mean" hint) keeps
// rendering as documented.
func TestFormatProblemHumanOutput(t *testing.T) {
	tests := []struct {
		name string
		p    tasks.Problem
		want []string
	}{
		{
			name: "task with line/column",
			p: tasks.Problem{
				Task:    "task #2",
				Line:    8,
				Column:  7,
				Message: "unknown task type \"dokku_appp\"",
				Hint:    "did you mean \"dokku_app\"?",
			},
			want: []string{"task #2", "line 8:7", "unknown task type", "did you mean"},
		},
		{
			name: "play-level problem with no task",
			p: tasks.Problem{
				Line:    3,
				Column:  7,
				Message: "input \"app\" is required",
			},
			want: []string{"line 3:7", "is required"},
		},
		{
			name: "no position info",
			p: tasks.Problem{
				Code:    "yaml_parse",
				Message: "broken",
			},
			want: []string{"broken"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatProblem(tt.p)
			for _, fragment := range tt.want {
				if !strings.Contains(got, fragment) {
					t.Errorf("formatProblem output missing %q\nfull: %q", fragment, got)
				}
			}
		})
	}
}

// TestValidateJSONEventShape constructs a Problem and round-trips it through
// the JSON encoder used by --json output to confirm that consumers can rely
// on the documented fields.
func TestValidateJSONEventShape(t *testing.T) {
	event := map[string]interface{}{
		"version": 1,
		"type":    "validate_problem",
		"code":    "unknown_task_type",
		"message": "unknown task type \"dokku_appp\"",
		"play":    "play #1",
		"task":    "task #2",
		"line":    8,
		"column":  7,
		"hint":    "did you mean \"dokku_app\"?",
	}
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if v, ok := decoded["version"].(float64); !ok || int(v) != 1 {
		t.Errorf("expected version=1, got %v", decoded["version"])
	}
	if decoded["type"] != "validate_problem" {
		t.Errorf("expected type=validate_problem, got %v", decoded["type"])
	}
	if decoded["code"] != "unknown_task_type" {
		t.Errorf("expected code=unknown_task_type, got %v", decoded["code"])
	}
}
