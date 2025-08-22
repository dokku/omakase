package main

import (
	"fmt"
	"io"
	"omakase/tasks"
	"strconv"

	sigil "github.com/gliderlabs/sigil"
	flag "github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"
)

type Argument struct {
	Required    bool
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
		"true": true,
		"yes":  true,
		"on":   true,
		"y":    true,
		"Y":    true,
	}
	return falseStrings[s]
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

func parseArgs(data []byte) (map[string]interface{}, error) {
	context := make(map[string]interface{})
	inputs, err := getInputVariables(data)
	if err != nil {
		return context, err
	}

	inputs["tasks"] = &tasks.Input{
		Name:        "tasks",
		Default:     "tasks.yml",
		Description: "a yaml file containing a task list",
	}

	arguments := make(map[string]*Argument)
	for _, input := range inputs {
		arg := Argument{Required: input.Required}
		switch input.Type {
		case "string", "":
			arg.SetStringValue(flag.String(input.Name, input.Default, input.Description))
		case "int":
			i, err := strconv.Atoi(input.Default)
			if err != nil {
				return context, fmt.Errorf("Error parsing input '%s': %v", input.Name, err.Error())
			}
			arg.SetIntValue(flag.Int(input.Name, i, input.Description))
		case "float":
			f, err := strconv.ParseFloat(input.Default, 64)
			if err != nil {
				return context, fmt.Errorf("Error parsing input '%s': %v", input.Name, err.Error())
			}
			arg.SetFloatValue(flag.Float64(input.Name, f, input.Description))
		case "bool":
			if isTrueString(input.Default) {
				arg.SetBoolValue(flag.Bool(input.Name, true, input.Description))
			} else if isFalseString(input.Default) {
				arg.SetBoolValue(flag.Bool(input.Name, false, input.Description))
			} else {
				return context, fmt.Errorf("Error parsing input '%s': invalid default value", input.Name)
			}
		default:
			return context, fmt.Errorf("Error parsing input '%s': invalid type", input.Name)
		}
		arguments[input.Name] = &arg
	}

	flag.Parse()

	for name, argument := range arguments {
		if argument.Required && !argument.HasValue() {
			return context, fmt.Errorf("Missing flag '--%s'", name)
		}
		context[name] = argument.GetValue()
	}

	return context, nil
}

func parseInputYaml(data []byte) (map[string]*tasks.Input, error) { // read variables and ensure they all exist
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
