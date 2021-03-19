package main

import (
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
	sigil "github.com/gliderlabs/sigil"
	_ "github.com/gliderlabs/sigil/builtin"
	yaml "gopkg.in/yaml.v2"
)

type Recipe []struct {
	Inputs []Input         `yaml:"inputs,omitempty"`
	Tasks  []TaskContainer `yaml:"tasks,omitempty"`
}

type Input struct {
	Name        string `yaml:"name"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

type TaskContainer struct {
	Name      string
	DokkuApp  *DokkuApp  `yaml:"dokku_app,omitempty"`
	DokkuSync *DokkuSync `yaml:"dokku_sync,omitempty"`
}

type Task interface {
	Execute(context struct{}) error
}

type DokkuApp struct {
	App   string `required:"true" yaml:"app"`
	State string `required:"true" yaml:"state" default:"present"`
}

type DokkuSync struct {
	App        string `required:"true" yaml:"app"`
	Repository string `required:"true" yaml:"repository"`
	State      string `required:"true" yaml:"state" default:"present"`
}

func parseInput(data []byte) (map[string]interface{}, []string, error) { // read variables and ensure they all exist
	context := make(map[string]interface{})
	required := []string{}
	t := Recipe{}
	if err := yaml.Unmarshal(data, &t); err != nil {
		return context, required, err
	}

	for _, recipe := range t {
		if len(recipe.Inputs) == 0 {
			continue
		}

		for _, input := range recipe.Inputs {
			context[input.Name] = input.Default
			if input.Required {
				required = append(required, input.Name)
			}
		}
	}

	return context, required, nil
}

func main() {
	data, err := ioutil.ReadFile("tasks.yml")
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	vars := make(map[string]interface{})
	render, err := sigil.Execute(data, vars, "tasks")
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	out, err := ioutil.ReadAll(&render)
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	context, requiredFields, err := parseInput(out)
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	// TODO: allow for custom path to task.yml
	flag.Parse()
	for _, arg := range flag.Args() {
		parts := strings.SplitN(arg, "=", 2)
		println(parts)
		if len(parts) == 2 {
			if _, ok := context[parts[0]]; ok {
				context[parts[0]] = parts[1]
			}
		}
	}

	for _, requiredField := range requiredFields {
		if value := context[requiredField]; value == "" {
			log.Fatalf("Required value for %s, found none", requiredField)
		}
	}

	render, err = sigil.Execute(data, context, "tasks")
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	out, err = ioutil.ReadAll(&render)
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	recipe := Recipe{}
	if err := yaml.Unmarshal([]byte(out), &recipe); err != nil {
		log.Fatalf("error: %v", err.Error())
	}

	tasks := []Task{}
	for _, t := range recipe[0].Tasks {
		// TODO: handle multiple contexts
		if t.DokkuApp != nil {
			tasks = append(tasks, AppTask{context: *t.DokkuApp})
			continue
		}
		if t.DokkuSync != nil {
			tasks = append(tasks, SyncTask{context: *t.DokkuSync})
			continue
		}
	}

	spew.Dump(tasks)

	return
}
