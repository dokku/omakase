package tasks

import (
	"encoding/json"
	"fmt"
	"docket/subprocess"
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

// GetName returns the name of the example
func (e GitSyncTaskExample) GetName() string {
	return e.Name
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
	return MarshalExamples([]GitSyncTaskExample{
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
	})
}

// Execute syncs a git repository to a dokku application
func (t GitSyncTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		"present": func() TaskOutputState { return syncGitRepository(t) },
	})
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
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// init registers the GitSyncTask with the task registry
func init() {
	RegisterTask(&GitSyncTask{})
}
