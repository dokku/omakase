package commands

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dokku/docket/tasks"

	sigil "github.com/gliderlabs/sigil"
	flag "github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v3"
)

type Argument struct {
	Required    bool
	Sensitive   bool
	boolValue   *bool
	floatValue  *float64
	intValue    *int
	stringValue *string
}

func (c Argument) GetValue() interface{} {
	if c.boolValue != nil {
		return c.boolValue
	} else if c.intValue != nil {
		return c.intValue
	} else if c.floatValue != nil {
		return c.floatValue
	} else if c.stringValue != nil && *c.stringValue != "" {
		return c.stringValue
	}
	return nil
}

func (c Argument) HasValue() bool {
	return c.GetValue() != nil
}

// StringValue returns the argument's value formatted as the same string sigil
// will substitute into the rendered YAML. Returns "" when no value is set.
// Used to register sensitive input values with the subprocess masker.
func (c Argument) StringValue() string {
	switch v := c.GetValue().(type) {
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case *int:
		if v == nil {
			return ""
		}
		return strconv.Itoa(*v)
	case *float64:
		if v == nil {
			return ""
		}
		return strconv.FormatFloat(*v, 'g', -1, 64)
	case *bool:
		if v == nil {
			return ""
		}
		return strconv.FormatBool(*v)
	}
	return ""
}

func (c *Argument) SetBoolValue(ptr *bool) {
	c.boolValue = ptr
}

func (c *Argument) SetFloatValue(ptr *float64) {
	c.floatValue = ptr
}

func (c *Argument) SetIntValue(ptr *int) {
	c.intValue = ptr
}

func (c *Argument) SetStringValue(ptr *string) {
	c.stringValue = ptr
}

func isTrueString(s string) bool {
	trueStrings := map[string]bool{
		"true": true,
		"yes":  true,
		"on":   true,
		"y":    true,
		"Y":    true,
	}
	return trueStrings[s]
}

func isFalseString(s string) bool {
	falseStrings := map[string]bool{
		"false": true,
		"no":    true,
		"off":   true,
		"n":     true,
		"N":     true,
	}
	return falseStrings[s]
}

func getTaskYamlFilename(s []string) string {
	for i, arg := range s {
		if arg == "--tasks" {
			if len(s) > i+1 {
				return s[i+1]
			}
		}
		if taskFile, found := strings.CutPrefix(arg, "--tasks="); found {
			return taskFile
		}
	}
	return "tasks.yml"
}

func getInputVariables(data []byte) (map[string]*tasks.Input, error) {
	vars := make(map[string]interface{})
	render, err := sigil.Execute(data, vars, "tasks")
	if err != nil {
		return map[string]*tasks.Input{}, fmt.Errorf("sigil error: %v", err.Error())
	}

	out, err := io.ReadAll(&render)
	if err != nil {
		return map[string]*tasks.Input{}, fmt.Errorf("render error: %v", err.Error())
	}

	return parseInputYaml(out)
}

// registerInputFlags reads the task file inputs and registers a flag for each
// dynamic input on the given FlagSet. It returns the Argument map keyed by
// input name so the caller can collect values after flags.Parse.
func registerInputFlags(f *flag.FlagSet, data []byte) (map[string]*Argument, error) {
	arguments := make(map[string]*Argument)
	inputs, err := getInputVariables(data)
	if err != nil {
		return arguments, err
	}

	for _, input := range inputs {
		if input.Name == "tasks" {
			continue
		}
		arg := &Argument{Required: input.Required, Sensitive: input.Sensitive}
		switch input.Type {
		case "string", "":
			arg.SetStringValue(f.String(input.Name, input.Default, input.Description))
		case "int":
			i, err := strconv.Atoi(input.Default)
			if err != nil {
				return arguments, fmt.Errorf("Error parsing input '%s': %v", input.Name, err.Error())
			}
			arg.SetIntValue(f.Int(input.Name, i, input.Description))
		case "float":
			ff, err := strconv.ParseFloat(input.Default, 64)
			if err != nil {
				return arguments, fmt.Errorf("Error parsing input '%s': %v", input.Name, err.Error())
			}
			arg.SetFloatValue(f.Float64(input.Name, ff, input.Description))
		case "bool":
			if isTrueString(input.Default) {
				arg.SetBoolValue(f.Bool(input.Name, true, input.Description))
			} else if isFalseString(input.Default) {
				arg.SetBoolValue(f.Bool(input.Name, false, input.Description))
			} else {
				return arguments, fmt.Errorf("Error parsing input '%s': invalid default value", input.Name)
			}
		default:
			return arguments, fmt.Errorf("Error parsing input '%s': invalid type", input.Name)
		}
		arguments[input.Name] = arg
	}

	return arguments, nil
}

func parseInputYaml(data []byte) (map[string]*tasks.Input, error) {
	inputs := make(map[string]*tasks.Input)
	t := tasks.Recipe{}
	if err := yaml.Unmarshal(data, &t); err != nil {
		return inputs, err
	}

	for _, recipe := range t {
		if len(recipe.Inputs) == 0 {
			continue
		}

		for name := range recipe.Inputs {
			input := recipe.Inputs[name]
			inputs[input.Name] = &input
		}
	}

	return inputs, nil
}
