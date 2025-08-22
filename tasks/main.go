package tasks

import (
	"crypto/rand"
	"fmt"
	"io"
	"reflect"
	"strings"

	sigil "github.com/gliderlabs/sigil"
	"github.com/gobuffalo/flect"
	jsoniter "github.com/json-iterator/go"
	yaml "gopkg.in/yaml.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Recipe []struct {
	Inputs []Input                  `yaml:"inputs,omitempty"`
	Tasks  []map[string]interface{} `yaml:"tasks,omitempty"`
}

type Input struct {
	Name        string `yaml:"name"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Type        string `yaml:"type"`
	value       string
}

type Tasks struct {
	Name        string
	AppTask     *AppTask     `yaml:"dokku_app,omitempty"`
	GitSyncTask *GitSyncTask `yaml:"dokku_sync,omitempty"`
}

type TaskOutputState struct {
	Changed bool
	Error   error
	Message string
	Meta    struct{}
	State   string
}

type Task interface {
	DesiredState() string
	Execute() TaskOutputState
}

// Global registry for Tasks.
var registeredTasks map[string]Task

func RegisterTask(t Task) {
	if len(registeredTasks) == 0 {
		registeredTasks = make(map[string]Task)
	}

	var name string
	if t := reflect.TypeOf(t); t.Kind() == reflect.Ptr {
		name = "*" + t.Elem().Name()
	} else {
		name = t.Name()
	}

	name = flect.Underscore(name)
	registeredTasks[fmt.Sprintf("dokku_%s", strings.TrimSuffix(name, "_task"))] = t
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

func GetTasks(data []byte, context map[string]interface{}) (OrderedStringTaskMap, error) {
	tasks := OrderedStringTaskMap{}
	render, err := sigil.Execute(data, context, "tasks")
	if err != nil {
		return tasks, fmt.Errorf("re-render error: %v", err.Error())
	}

	out, err := io.ReadAll(&render)
	if err != nil {
		return tasks, fmt.Errorf("read error: %v", err.Error())
	}

	recipe := Recipe{}
	if err := yaml.Unmarshal([]byte(out), &recipe); err != nil {
		return tasks, fmt.Errorf("unmarshal error: %v", err.Error())
	}

	i := 0
	validTasks := make([]string, len(registeredTasks))
	for k := range registeredTasks {
		validTasks[i] = k
		i++
	}

	for i, t := range recipe[0].Tasks {
		if len(t) > 2 {
			return tasks, fmt.Errorf("task parse error: task #%d has too many properties - expected=2 actual=%d", i+1, len(t))
		}

		name, ok := t["name"]
		if len(t) == 2 && !ok {
			keys := make([]string, len(t))

			j := 0
			for k := range t {
				keys[j] = k
				j++
			}

			return tasks, fmt.Errorf("task parse error: task #%d has an unexpected property - properties=%v", i+1, keys)
		}

		if !ok {
			b := make([]byte, 8)
			if _, err := rand.Read(b); err != nil {
				return tasks, fmt.Errorf("task parse error: task #%d had no task name and there was a failure to generate random task name - %s", i+1, err)
			}
			name = fmt.Sprintf("task #%d %X", i+1, b)
		}

		detected := false
		for taskName, registeredTask := range registeredTasks {
			config, ok := t[taskName]
			if !ok {
				continue
			}

			marshaled, err := json.Marshal(config)
			if err != nil {
				return tasks, fmt.Errorf("task parse error: task #%d failed to marshal config to json - %s", i+1, err)
			}

			v := reflect.New(reflect.TypeOf(registeredTask))
			if err := json.Unmarshal([]byte(marshaled), v.Interface()); err != nil {
				return tasks, fmt.Errorf("task parse error: task #%d failed to decode to %s - %s", i+1, taskName, err)
			}

			task := v.Elem().Interface().(Task)
			tasks.Set(name.(string), task)
			detected = true
			break
		}

		if !detected {
			return tasks, fmt.Errorf("task parse error: task #%d was not a valid task - valid_tasks=%v", i+1, validTasks)
		}
	}

	return tasks, nil
}
