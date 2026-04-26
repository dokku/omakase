package commands

import (
	"strings"
	"testing"

	"github.com/dokku/docket/tasks"
)

func TestIsTrueString(t *testing.T) {
	trueValues := []string{"true", "yes", "on", "y", "Y"}
	for _, v := range trueValues {
		if !isTrueString(v) {
			t.Errorf("isTrueString(%q) = false, want true", v)
		}
	}

	falseValues := []string{"false", "no", "off", "n", "N", "", "maybe", "1", "0"}
	for _, v := range falseValues {
		if isTrueString(v) {
			t.Errorf("isTrueString(%q) = true, want false", v)
		}
	}
}

func TestIsFalseString(t *testing.T) {
	falseValues := []string{"false", "no", "off", "n", "N"}
	for _, v := range falseValues {
		if !isFalseString(v) {
			t.Errorf("isFalseString(%q) = false, want true", v)
		}
	}

	trueValues := []string{"true", "yes", "on", "y", "Y", "", "maybe", "1", "0"}
	for _, v := range trueValues {
		if isFalseString(v) {
			t.Errorf("isFalseString(%q) = true, want false", v)
		}
	}
}

func TestIsTrueAndFalseAreMutuallyExclusive(t *testing.T) {
	allValues := []string{"true", "false", "yes", "no", "on", "off", "y", "n", "Y", "N", "", "maybe"}
	for _, v := range allValues {
		isTrue := isTrueString(v)
		isFalse := isFalseString(v)
		if isTrue && isFalse {
			t.Errorf("value %q is both true and false", v)
		}
	}
}

func TestGetTaskYamlFilename(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "default when no --tasks flag",
			args:     []string{"docket"},
			expected: "tasks.yml",
		},
		{
			name:     "separate --tasks flag",
			args:     []string{"docket", "--tasks", "custom.yml"},
			expected: "custom.yml",
		},
		{
			name:     "equals --tasks=flag",
			args:     []string{"docket", "--tasks=custom.yml"},
			expected: "custom.yml",
		},
		{
			name:     "--tasks at end with no value",
			args:     []string{"docket", "--tasks"},
			expected: "tasks.yml",
		},
		{
			name:     "--tasks with other flags before",
			args:     []string{"docket", "--app", "myapp", "--tasks", "other.yml"},
			expected: "other.yml",
		},
		{
			name:     "uses passed parameter not os.Args",
			args:     []string{"anything", "--tasks", "from-param.yml"},
			expected: "from-param.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTaskYamlFilename(tt.args)
			if result != tt.expected {
				t.Errorf("getTaskYamlFilename(%v) = %q, want %q", tt.args, result, tt.expected)
			}
		})
	}
}

func TestArgumentGetValue(t *testing.T) {
	t.Run("bool value", func(t *testing.T) {
		b := true
		arg := Argument{}
		arg.SetBoolValue(&b)
		if arg.GetValue() == nil {
			t.Error("GetValue() returned nil for bool argument")
		}
		if !arg.HasValue() {
			t.Error("HasValue() returned false for bool argument")
		}
	})

	t.Run("int value", func(t *testing.T) {
		i := 42
		arg := Argument{}
		arg.SetIntValue(&i)
		if arg.GetValue() == nil {
			t.Error("GetValue() returned nil for int argument")
		}
		if !arg.HasValue() {
			t.Error("HasValue() returned false for int argument")
		}
	})

	t.Run("float value", func(t *testing.T) {
		f := 3.14
		arg := Argument{}
		arg.SetFloatValue(&f)
		if arg.GetValue() == nil {
			t.Error("GetValue() returned nil for float argument")
		}
		if !arg.HasValue() {
			t.Error("HasValue() returned false for float argument")
		}
	})

	t.Run("string value non-empty", func(t *testing.T) {
		s := "hello"
		arg := Argument{}
		arg.SetStringValue(&s)
		if arg.GetValue() == nil {
			t.Error("GetValue() returned nil for non-empty string argument")
		}
		if !arg.HasValue() {
			t.Error("HasValue() returned false for non-empty string argument")
		}
	})

	t.Run("string value empty", func(t *testing.T) {
		s := ""
		arg := Argument{}
		arg.SetStringValue(&s)
		if arg.GetValue() != nil {
			t.Error("GetValue() returned non-nil for empty string argument")
		}
		if arg.HasValue() {
			t.Error("HasValue() returned true for empty string argument")
		}
	})

	t.Run("no value set", func(t *testing.T) {
		arg := Argument{}
		if arg.GetValue() != nil {
			t.Error("GetValue() returned non-nil for unset argument")
		}
		if arg.HasValue() {
			t.Error("HasValue() returned true for unset argument")
		}
	})
}

func TestParseInputYamlValidInputs(t *testing.T) {
	data := []byte(`---
- inputs:
    - name: app
      default: "myapp"
      description: "Application name"
      required: true
      type: string
    - name: port
      default: "8080"
      description: "Port number"
      type: int
  tasks: []
`)
	inputs, err := parseInputYaml(data)
	if err != nil {
		t.Fatalf("parseInputYaml failed: %v", err)
	}

	if len(inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(inputs))
	}

	app, ok := inputs["app"]
	if !ok {
		t.Fatal("expected 'app' input")
	}
	if app.Default != "myapp" {
		t.Errorf("app.Default = %q, want %q", app.Default, "myapp")
	}
	if !app.Required {
		t.Error("app.Required = false, want true")
	}
	if app.Type != "string" {
		t.Errorf("app.Type = %q, want %q", app.Type, "string")
	}

	port, ok := inputs["port"]
	if !ok {
		t.Fatal("expected 'port' input")
	}
	if port.Default != "8080" {
		t.Errorf("port.Default = %q, want %q", port.Default, "8080")
	}
	if port.Type != "int" {
		t.Errorf("port.Type = %q, want %q", port.Type, "int")
	}
}

func TestParseInputYamlNoInputs(t *testing.T) {
	data := []byte("---\n- tasks: []\n")
	inputs, err := parseInputYaml(data)
	if err != nil {
		t.Fatalf("parseInputYaml failed: %v", err)
	}
	if len(inputs) != 0 {
		t.Errorf("expected 0 inputs, got %d", len(inputs))
	}
}

func TestParseInputYamlInvalidYaml(t *testing.T) {
	data := []byte("not valid yaml: [[[")
	_, err := parseInputYaml(data)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseInputYamlAllTypes(t *testing.T) {
	data := []byte(`---
- inputs:
    - name: str_input
      type: string
      default: "hello"
    - name: int_input
      type: int
      default: "42"
    - name: float_input
      type: float
      default: "3.14"
    - name: bool_input
      type: bool
      default: "true"
  tasks: []
`)
	inputs, err := parseInputYaml(data)
	if err != nil {
		t.Fatalf("parseInputYaml failed: %v", err)
	}

	tests := []struct {
		name     string
		wantType string
	}{
		{"str_input", "string"},
		{"int_input", "int"},
		{"float_input", "float"},
		{"bool_input", "bool"},
	}

	for _, tt := range tests {
		input, ok := inputs[tt.name]
		if !ok {
			t.Errorf("expected input %q", tt.name)
			continue
		}
		if input.Type != tt.wantType {
			t.Errorf("input %q type = %q, want %q", tt.name, input.Type, tt.wantType)
		}
	}
}

func TestParseInputYamlMultipleRecipes(t *testing.T) {
	data := []byte(`---
- inputs:
    - name: first
      default: "a"
  tasks: []
- inputs:
    - name: second
      default: "b"
  tasks: []
`)
	inputs, err := parseInputYaml(data)
	if err != nil {
		t.Fatalf("parseInputYaml failed: %v", err)
	}

	if _, ok := inputs["first"]; !ok {
		t.Error("expected 'first' input from first recipe section")
	}
	if _, ok := inputs["second"]; !ok {
		t.Error("expected 'second' input from second recipe section")
	}
}

func TestGetInputVariablesValid(t *testing.T) {
	data := []byte(`---
- inputs:
    - name: app
      default: "myapp"
      description: "App name"
      required: true
  tasks: []
`)
	inputs, err := getInputVariables(data)
	if err != nil {
		t.Fatalf("getInputVariables failed: %v", err)
	}

	app, ok := inputs["app"]
	if !ok {
		t.Fatal("expected 'app' input")
	}
	if app.Default != "myapp" {
		t.Errorf("app.Default = %q, want %q", app.Default, "myapp")
	}
	if !app.Required {
		t.Error("app.Required = false, want true")
	}
}

func TestGetInputVariablesTemplateError(t *testing.T) {
	data := []byte(`---
- inputs:
    - name: {{ .broken
  tasks: []
`)
	_, err := getInputVariables(data)
	if err == nil {
		t.Fatal("expected error for bad template syntax")
	}
	if !strings.Contains(err.Error(), "sigil error") {
		t.Errorf("expected 'sigil error', got: %v", err)
	}
}

func TestInputSetValueAndGetValue(t *testing.T) {
	input := tasks.Input{}
	err := input.SetValue("hello")
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}
	if input.GetValue() != "hello" {
		t.Errorf("GetValue() = %q, want %q", input.GetValue(), "hello")
	}
	if !input.HasValue() {
		t.Error("HasValue() = false, want true")
	}
}

func TestInputHasValueEmpty(t *testing.T) {
	input := tasks.Input{}
	if input.HasValue() {
		t.Error("HasValue() = true for unset input, want false")
	}
	if input.GetValue() != "" {
		t.Errorf("GetValue() = %q for unset input, want empty", input.GetValue())
	}
}

func TestInputSetValueOverwrite(t *testing.T) {
	input := tasks.Input{}
	input.SetValue("first")
	input.SetValue("second")
	if input.GetValue() != "second" {
		t.Errorf("GetValue() = %q after overwrite, want %q", input.GetValue(), "second")
	}
}
