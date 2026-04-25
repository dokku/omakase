//go:generate go run docs.go docs
package main

import (
	"fmt"
	"log"
	"docket/tasks"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	docsFolderName := "../" + os.Args[1]
	// expand docsFolderName
	docsFolderName, err := filepath.Abs(docsFolderName)
	if err != nil {
		log.Fatalf("failed to expand docs folder name: %v", err)
	}

	// create docs folder if it doesn't exist
	if _, err := os.Stat(docsFolderName); os.IsNotExist(err) {
		err = os.MkdirAll(docsFolderName, 0755)
		if err != nil {
			log.Fatalf("failed to create docs folder: %v", err)
		}
	}

	markdownTemplate := `
# %s

%s
%s
`

	sectionTemplate := `
## %s

%syaml
%s
%s`

	codefence := "```"

	// read in all registered tasks
	registeredTasks := tasks.RegisteredTasks

	// for each registered task, generate a docs file
	for taskName, task := range registeredTasks {
		fmt.Println(taskName)

		examples, err := task.Examples()
		if err != nil {
			log.Fatalf("failed to get examples for task %s: %v", taskName, err)
		}

		docblock := task.Doc()

		var exampleSections []string
		for _, example := range examples {
			example := fmt.Sprintf(sectionTemplate, example.Name, codefence, strings.TrimSpace(example.Codeblock), codefence)
			exampleSections = append(exampleSections, example)
		}

		examplesYaml := strings.Join(exampleSections, "\n")
		markdown := fmt.Sprintf(markdownTemplate, taskName, docblock, examplesYaml)
		output := strings.TrimSpace(markdown) + "\n"

		taskDocsFile := filepath.Join(docsFolderName, taskName+".md")
		err = os.WriteFile(taskDocsFile, []byte(output), 0644)
		if err != nil {
			log.Fatalf("failed to write docblock: %v", err)
		}
	}
}
