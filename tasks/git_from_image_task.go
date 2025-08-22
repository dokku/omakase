package tasks

import "omakase/subprocess"

// git:from-image [--build-dir DIRECTORY] <app> <docker-image> [<git-username> <git-email>]

type GitFromImageTask struct {
	App         string `required:"true" yaml:"app"`
	Image       string `required:"true" yaml:"image"`
	BuildDir    string `required:"false" yaml:"build_dir"`
	GitUsername string `required:"false" yaml:"git_username"`
	GitEmail    string `required:"false" yaml:"git_email"`
	State       string `required:"true" yaml:"state" default:"deployed"`
}

func (t GitFromImageTask) DesiredState() string {
	return t.State
}

func (t GitFromImageTask) Execute() TaskOutputState {
	funcMap := map[string]func(GitFromImageTask) TaskOutputState{
		"deploy": deployGitFromImage,
	}

	fn := funcMap[t.State]
	return fn(t)
}

func deployGitFromImage(t GitFromImageTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "undeployed",
	}

	// todo: get current deployed image and short-circuit if it matches the desired image

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

func init() {
	RegisterTask(&GitFromImageTask{})
}
