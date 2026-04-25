package tasks

import (
	"encoding/json"
	"fmt"
	"docket/subprocess"

	yaml "gopkg.in/yaml.v3"
)

// GitSyncTask syncs a git repository to a dokku application
type GitSyncTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Remote is the git remote url to sync
	Remote string `required:"true" yaml:"remote"`

	// GitRef is the git reference to sync
	GitRef string `required:"false" yaml:"git_ref"`

	// Build triggers an application build after syncing
	Build bool `required:"false" yaml:"build"`

	// BuildIfChanges triggers a build only if changes are detected
	BuildIfChanges bool `required:"false" yaml:"build_if_changes"`

	// SkipDeployBranch skips automatically setting the deploy-branch property
	SkipDeployBranch bool `required:"false" yaml:"skip_deploy_branch"`

	// State is the desired state of the git sync
	State State `required:"false" yaml:"state" default:"present" options:"present"`
}

// GitSyncTaskExample contains an example of a GitSyncTask
type GitSyncTaskExample struct {
	// Name is the task name holding the GitSyncTask description
	Name string `yaml:"-"`

	// GitSyncTask is the GitSyncTask configuration
	GitSyncTask GitSyncTask `yaml:"dokku_git_sync"`
}

// DesiredState returns the desired state of the git sync
func (t GitSyncTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the git sync task
func (t GitSyncTask) Doc() string {
	return "Syncs a git repository to a dokku application"
}

// Examples returns the examples for the git sync task
func (t GitSyncTask) Examples() ([]Doc, error) {
	examples := []GitSyncTaskExample{
		{
			Name: "Sync a git repository to an app",
			GitSyncTask: GitSyncTask{
				App:    "hello-world",
				Remote: "https://github.com/dokku/smoke-test-app.git",
			},
		},
		{
			Name: "Sync a git repository with a specific ref and build",
			GitSyncTask: GitSyncTask{
				App:    "hello-world",
				Remote: "https://github.com/dokku/smoke-test-app.git",
				GitRef: "main",
				Build:  true,
			},
		},
	}

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

// Execute syncs a git repository to a dokku application
func (t GitSyncTask) Execute() TaskOutputState {
	funcMap := map[State]func(GitSyncTask) TaskOutputState{
		"present": syncGitRepository,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn(t)
}

// checkAppSyncState checks if the app is already synced from the expected remote and ref
func checkAppSyncState(app, expectedRemote, expectedRef string) bool {
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

	expectedMetadata := fmt.Sprintf("%s#%s", expectedRemote, expectedRef)
	return source.Source == "git-sync" && source.SourceMetadata == expectedMetadata
}

// syncGitRepository syncs a git repository to a dokku application
func syncGitRepository(t GitSyncTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	if checkAppSyncState(t.App, t.Remote, t.GitRef) {
		state.State = "present"
		return state
	}

	args := []string{
		"git:sync",
	}

	if t.Build {
		args = append(args, "--build")
	}
	if t.BuildIfChanges {
		args = append(args, "--build-if-changes")
	}
	if t.SkipDeployBranch {
		args = append(args, "--skip-deploy-branch")
	}

	args = append(args, t.App, t.Remote)

	if t.GitRef != "" {
		args = append(args, t.GitRef)
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
	state.State = "present"
	return state
}

// init registers the GitSyncTask with the task registry
func init() {
	RegisterTask(&GitSyncTask{})
}
