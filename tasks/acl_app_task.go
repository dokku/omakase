package tasks

import (
	"fmt"
	"strings"

	"docket/subprocess"
)

// AclAppTask manages the dokku-acl access list for a dokku application
type AclAppTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Users is the list of users to add or remove from the ACL
	Users []string `required:"false" yaml:"users"`

	// State is the desired state of the ACL entries
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// AclAppTaskExample contains an example of an AclAppTask
type AclAppTaskExample struct {
	// Name is the task name holding the AclAppTask description
	Name string `yaml:"-"`

	// AclAppTask is the AclAppTask configuration
	AclAppTask AclAppTask `yaml:"dokku_acl_app"`
}

// GetName returns the name of the example
func (e AclAppTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the acl app task
func (t AclAppTask) Doc() string {
	return "Manages the dokku-acl access list for a dokku application"
}

// Examples returns the examples for the acl app task
func (t AclAppTask) Examples() ([]Doc, error) {
	return MarshalExamples([]AclAppTaskExample{
		{
			Name: "Grant users access to an app",
			AclAppTask: AclAppTask{
				App:   "node-js-app",
				Users: []string{"alice", "bob"},
			},
		},
		{
			Name: "Revoke a user's access to an app",
			AclAppTask: AclAppTask{
				App:   "node-js-app",
				Users: []string{"bob"},
				State: StateAbsent,
			},
		},
		{
			Name: "Clear the entire ACL for an app",
			AclAppTask: AclAppTask{
				App:   "node-js-app",
				State: StateAbsent,
			},
		},
	})
}

// Execute manages the app ACL
func (t AclAppTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		StatePresent: func() TaskOutputState { return addAclAppUsers(t) },
		StateAbsent:  func() TaskOutputState { return removeAclAppUsers(t) },
	})
}

// getAclAppUsers reads the current ACL for an app via `acl:list APP`. The
// plugin emits one username per line; an empty ACL produces no output.
func getAclAppUsers(app string) (map[string]bool, error) {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "acl:list", app},
	})
	if err != nil {
		return nil, err
	}

	users := map[string]bool{}
	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		users[trimmed] = true
	}
	return users, nil
}

// addAclAppUsers adds users to an app's ACL, skipping ones already present
func addAclAppUsers(t AclAppTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if t.App == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}
	if len(t.Users) == 0 {
		state.Error = fmt.Errorf("'users' must not be empty for state 'present'")
		return state
	}

	current, err := getAclAppUsers(t.App)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	var toAdd []string
	for _, user := range t.Users {
		if !current[user] {
			toAdd = append(toAdd, user)
		}
	}

	if len(toAdd) == 0 {
		state.State = StatePresent
		return state
	}

	for _, user := range toAdd {
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "acl:add", t.App, user},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
	}

	state.Changed = true
	state.State = StatePresent
	return state
}

// removeAclAppUsers removes users from an app's ACL. With an empty Users list,
// removes every entry currently in the ACL (i.e. clears it).
func removeAclAppUsers(t AclAppTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StatePresent,
	}

	if t.App == "" {
		state.Error = fmt.Errorf("'app' is required")
		return state
	}

	current, err := getAclAppUsers(t.App)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	var toRemove []string
	if len(t.Users) == 0 {
		for user := range current {
			toRemove = append(toRemove, user)
		}
	} else {
		for _, user := range t.Users {
			if current[user] {
				toRemove = append(toRemove, user)
			}
		}
	}

	if len(toRemove) == 0 {
		state.State = StateAbsent
		return state
	}

	for _, user := range toRemove {
		result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
			Command: "dokku",
			Args:    []string{"--quiet", "acl:remove", t.App, user},
		})
		if err != nil {
			return TaskOutputErrorFromExec(state, err, result)
		}
	}

	state.Changed = true
	state.State = StateAbsent
	return state
}

// init registers the AclAppTask with the task registry
func init() {
	RegisterTask(&AclAppTask{})
}
