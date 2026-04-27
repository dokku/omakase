package tasks

import (
	"fmt"

	"github.com/dokku/docket/subprocess"
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

// GetName returns the name of the example
func (e GitFromImageTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the git from image task
func (t GitFromImageTask) Doc() string {
	return "Deploys a git repository from a docker image"
}

// Examples returns the examples for the git from image task
func (t GitFromImageTask) Examples() ([]Doc, error) {
	return MarshalExamples([]GitFromImageTaskExample{})
}

// Execute deploys a git repository from a docker image
func (t GitFromImageTask) Execute() TaskOutputState {
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the GitFromImageTask would produce.
func (t GitFromImageTask) Plan() PlanResult {
	return DispatchPlan(t.State, map[State]func() PlanResult{
		StateDeployed: func() PlanResult {
			if checkAppSourceImage(t.App, t.Image) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusModify,
				Reason:    "image source drift",
				Mutations: []string{fmt.Sprintf("git:from-image %s %s", t.App, t.Image)},
				apply: func() TaskOutputState {
					state := TaskOutputState{Changed: false, State: "undeployed"}
					args := []string{"git:from-image"}
					if t.BuildDir != "" {
						args = append(args, "--build-dir", t.BuildDir)
					}
					args = append(args, t.App, t.Image)
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
					state.Commands = append(state.Commands, result.Command)
					if err != nil {
						return TaskOutputErrorFromExec(state, err, result)
					}
					state.Changed = true
					state.State = StateDeployed
					return state
				},
			}
		},
	})
}

// checkAppSourceImage checks if the app is already deployed from a docker image
func checkAppSourceImage(app, expectedImage string) bool {
	source, err := getAppDeploySource(app)
	if err != nil {
		return false
	}

	return source.Source == "docker-image" && source.SourceMetadata == expectedImage
}

// init registers the GitFromImageTask with the task registry
func init() {
	RegisterTask(&GitFromImageTask{})
}
