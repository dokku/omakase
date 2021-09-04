package main

import (
	"io/ioutil"
	"log"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/gliderlabs/sigil/builtin"
)

func main() {
	// TODO: allow for custom path to task.yml
	data, err := ioutil.ReadFile("tasks.yml")
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
