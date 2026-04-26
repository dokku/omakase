package main

import (
	"log"
	"docket/tasks"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/gliderlabs/sigil/builtin"
)

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

func main() {
	taskFile := getTaskYamlFilename(os.Args)
	data, err := os.ReadFile(taskFile)
	if err != nil {
		log.Fatalf("read error: %v", err)
	}

	context, err := parseArgs(data)
	if err != nil {
		log.Fatalf("arg error: %v", err)
	}

	tasks, err := tasks.GetTasks(data, context)
	if err != nil {
		log.Fatalf("task error: %v", err)
	}

	spew.Dump(tasks)
	for _, name := range tasks.Keys() {
		task := tasks.Get(name)
		log.Printf("executing %s", name)
		state := task.Execute()
		if state.Error != nil {
			log.Fatalf("execute error: %v", state.Error)
		}

		if state.State != state.DesiredState {
			log.Fatalf("error: Invalid state found, expected=%v actual=%v", state.DesiredState, state.State)
		}
	}
}
