package main

import (
	"testing"
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
			args:     []string{"omakase"},
			expected: "tasks.yml",
		},
		{
			name:     "separate --tasks flag",
			args:     []string{"omakase", "--tasks", "custom.yml"},
			expected: "custom.yml",
		},
		{
			name:     "equals --tasks=flag",
			args:     []string{"omakase", "--tasks=custom.yml"},
			expected: "custom.yml",
		},
		{
			name:     "--tasks at end with no value",
			args:     []string{"omakase", "--tasks"},
			expected: "tasks.yml",
		},
		{
			name:     "--tasks with other flags before",
			args:     []string{"omakase", "--app", "myapp", "--tasks", "other.yml"},
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
