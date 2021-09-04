package main

import (
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/fatih/structtag"
	sigil "github.com/gliderlabs/sigil"
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
	Type        string `yaml:"type"`
	value       string
}

type TaskContainer struct {
	Name      string
	DokkuApp  *AppTask  `yaml:"dokku_app,omitempty"`
	DokkuSync *SyncTask `yaml:"dokku_sync,omitempty"`
}

type Task interface {
	DesiredState() string
	Execute() (string, error)
	NeedsExecution() bool
	SetDefaultDesiredState(state string)
}

func (i *Input) SetValue(value string) error {
	i.value = value
	return nil
}

func (i Input) HasValue() bool {
	return i.value != ""
}

func (i Input) GetValue() string {
	return i.value
}

func getTasks(data []byte, context map[string]interface{}) ([]Task, error) {
	tasks := []Task{}
	render, err := sigil.Execute(data, context, "tasks")
	if err != nil {
		return tasks, fmt.Errorf("re-render error: %v", err.Error())
	}

	out, err := ioutil.ReadAll(&render)
	if err != nil {
		return tasks, fmt.Errorf("read error: %v", err.Error())
	}

	recipe := Recipe{}
	if err := yaml.Unmarshal([]byte(out), &recipe); err != nil {
		return tasks, fmt.Errorf("unmarshal error: %v", err.Error())
	}

	for _, t := range recipe[0].Tasks {
		ts := map[interface{}]Task{
			AppTask{}:  t.DokkuApp,
			SyncTask{}: t.DokkuSync,
		}
		for i, task := range ts {
			if reflect.ValueOf(task).IsNil() {
				continue
			}

			defaultState, err := getDefaultState(i)
			if err != nil {
				return tasks, fmt.Errorf("task parse error: %v", err)
			}
			task.SetDefaultDesiredState(defaultState)
			tasks = append(tasks, task)
			continue
		}
	}

	return tasks, nil
}

func getDefaultState(i interface{}) (string, error) {
	state, _ := reflect.TypeOf(i).FieldByName("State")
	tags, err := structtag.Parse(string(state.Tag))
	if err != nil {
		return "", err
	}

	defaultState, err := tags.Get("default")
	if err != nil {
		return "", err
	}

	return defaultState.Name, nil
}
