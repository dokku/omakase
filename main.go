package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/gliderlabs/sigil/builtin"
)

func getTaskYamlFilename(s []string) string {
	for i, arg := range s {
		if arg == "--tasks" {
			if len(os.Args) > i {
				return os.Args[i+1]
			}
		}
		if strings.HasPrefix(arg, "--tasks=") {
			return strings.TrimPrefix(arg, "--tasks=")
		}
	}
	return "tasks.yml"
}

func main() {
	taskFile := getTaskYamlFilename(os.Args)
	data, err := ioutil.ReadFile(taskFile)
	if err != nil {
		log.Fatalf("read error: %v", err.Error())
	}

	context, err := parseArgs(data)
	if err != nil {
		log.Fatalf("arg error: %v", err.Error())
	}

	tasks, err := getTasks(data, context)
	if err != nil {
		log.Fatalf(err.Error())
	}

	spew.Dump(tasks)
	for _, task := range tasks {
		if !task.NeedsExecution() {
			continue
		}

		state, err := task.Execute()
		if err != nil {
			log.Fatalf("execute error: %v", err.Error())
		}

		if state != task.DesiredState() {
			log.Fatalf("error: Invalid state found, expected=%v actual=%v", task.DesiredState(), state)

		}
	}

	return
}
