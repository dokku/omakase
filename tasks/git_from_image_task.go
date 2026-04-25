package tasks

import (
	"encoding/json"
	"fmt"
	"docket/subprocess"

	yaml "gopkg.in/yaml.v3"
)

// git:from-image [--build-dir DIRECTORY] <app> <docker-image> [<git-username> <git-email>]

// GitFromImageTask deploys a git repository from a docker image
type GitFromImageTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Image is the docker image to deploy
	Image string `required:"true" yaml:"image"`

	// BuildDir is the directory to build the git repository
	BuildDir string `required:"false" yaml:"build_dir"`

	// GitUsername is the username to use for the git repository
	GitUsername string `required:"false" yaml:"git_username"`

	// GitEmail is the email to use for the git repository
	GitEmail string `required:"false" yaml:"git_email"`

	// State is the desired state of the git repository
	State State `required:"false" yaml:"state" default:"deployed" options:"deployed"`
}

// GitFromImageTaskExample contains an example of a GitFromImageTask
type GitFromImageTaskExample struct {
	// Name is the task name holding the GitFromImageTask description
	Name string `yaml:"-"`

	// GitFromImageTask is the GitFromImageTask configuration
	GitFromImageTask GitFromImageTask `yaml:"dokku_git_from_image"`
}

// DesiredState returns the desired state of the git repository
func (t GitFromImageTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the git from image task
func (t GitFromImageTask) Doc() string {
	return "Deploys a git repository from a docker image"
}

// Examples returns the examples for the builder property task
func (t GitFromImageTask) Examples() ([]Doc, error) {
	examples := []GitFromImageTaskExample{}

	var output []Doc
	for _, example := range examples {
		b, err := yaml.Marshal(example)
		if err != nil {
			return nil, err
		}

		output = append(output, Doc{
			Name:      example.Name,
			Codeblock: string(b),
		})
	}

	return output, nil
}

// Execute deploys a git repository from a docker image
func (t GitFromImageTask) Execute() TaskOutputState {
	funcMap := map[State]func(GitFromImageTask) TaskOutputState{
		"deployed": deployGitFromImage,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn(t)
}

// checkAppSourceImage checks if the app is already deployed from a docker image
func checkAppSourceImage(app, expectedImage string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"apps:report", app, "--format", "json"},
	})
	if err != nil {
		return false
	}

	type appSource struct {
		Source         string `json:"app-deploy-source"`
		SourceMetadata string `json:"app-deploy-source-metadata"`
	}

	var source appSource
	err = json.Unmarshal(result.StdoutBytes(), &source)
	if err != nil {
		return false
	}

	return source.Source == "docker-image" && source.SourceMetadata == expectedImage
}

// deployGitFromImage deploys a git repository from a docker image
func deployGitFromImage(t GitFromImageTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "undeployed",
	}

	if checkAppSourceImage(t.App, t.Image) {
		state.Changed = false
		state.State = "deployed"
		return state
	}

	args := []string{
		"git:from-image",
	}
	if t.BuildDir != "" {
		args = append(args, "--build-dir", t.BuildDir)
	}
	args = append(args, t.App, t.Image)

	// todo: ensure both the username and email are provided
	if t.GitUsername != "" {
		args = append(args, t.GitUsername)
	}
	if t.GitEmail != "" {
		args = append(args, t.GitEmail)
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		state.Error = err
		state.Message = result.StderrContents()
		return state
	}

	state.Changed = true
	state.State = "deployed"
	return state
}

// init registers the GitFromImageTask with the task registry
func init() {
	RegisterTask(&GitFromImageTask{})
}
