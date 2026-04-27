package tasks

import (
	"fmt"
	"github.com/dokku/docket/subprocess"
	"strings"
)

// HttpAuthTask manages HTTP authentication for a dokku application
type HttpAuthTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Username is the HTTP auth username
	Username string `required:"false" yaml:"username,omitempty"`

	// Password is the HTTP auth password
	Password string `required:"false" yaml:"password,omitempty"`

	// State is the state of the HTTP auth
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// HttpAuthTaskExample contains an example of an HttpAuthTask
type HttpAuthTaskExample struct {
	// Name is the task name holding the HttpAuthTask description
	Name string `yaml:"-"`

	// DokkuHttpAuth is the HttpAuthTask configuration
	DokkuHttpAuth HttpAuthTask `yaml:"dokku_http_auth"`
}

// GetName returns the name of the example
func (e HttpAuthTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the HTTP auth task
func (t HttpAuthTask) Doc() string {
	return "Manages HTTP authentication for a given dokku application"
}

// Examples returns a list of HttpAuthTaskExamples as yaml
func (t HttpAuthTask) Examples() ([]Doc, error) {
	return MarshalExamples([]HttpAuthTaskExample{
		{
			Name: "Enable HTTP authentication for an app",
			DokkuHttpAuth: HttpAuthTask{
				App:      "hello-world",
				Username: "admin",
				Password: "secret",
			},
		},
		{
			Name: "Disable HTTP authentication for an app",
			DokkuHttpAuth: HttpAuthTask{
				App:   "hello-world",
				State: "absent",
			},
		},
	})
}

// Execute enables or disables HTTP authentication for an app
func (t HttpAuthTask) Execute() TaskOutputState {
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the HttpAuthTask would produce.
func (t HttpAuthTask) Plan() PlanResult {
	if t.State == StatePresent && t.Username == "" {
		return PlanResult{Status: PlanStatusError, Error: fmt.Errorf("username is required when state is present")}
	}
	if t.State == StatePresent && t.Password == "" {
		return PlanResult{Status: PlanStatusError, Error: fmt.Errorf("password is required when state is present")}
	}
	return DispatchPlan(t.State, map[State]func() PlanResult{
		StatePresent: func() PlanResult {
			if httpAuthEnabled(t.App) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusCreate,
				Reason:    "http-auth disabled",
				Mutations: []string{"http-auth:on " + t.App},
				apply: func() TaskOutputState {
					state := TaskOutputState{Changed: false, State: StateAbsent}
					result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
						Command: "dokku",
						Args:    []string{"--quiet", "http-auth:on", t.App, t.Username, t.Password},
					})
					state.Commands = append(state.Commands, result.Command)
					if err != nil {
						return TaskOutputErrorFromExec(state, err, result)
					}
					state.Changed = true
					state.State = StatePresent
					return state
				},
			}
		},
		StateAbsent: func() PlanResult {
			if !httpAuthEnabled(t.App) {
				return PlanResult{InSync: true, Status: PlanStatusOK}
			}
			return PlanResult{
				InSync:    false,
				Status:    PlanStatusDestroy,
				Reason:    "http-auth enabled",
				Mutations: []string{"http-auth:off " + t.App},
				apply: func() TaskOutputState {
					state := TaskOutputState{Changed: false, State: StatePresent}
					result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
						Command: "dokku",
						Args:    []string{"--quiet", "http-auth:off", t.App},
					})
					state.Commands = append(state.Commands, result.Command)
					if err != nil {
						return TaskOutputErrorFromExec(state, err, result)
					}
					state.Changed = true
					state.State = StateAbsent
					return state
				},
			}
		},
	})
}

// httpAuthEnabled checks if HTTP authentication is enabled for an app
func httpAuthEnabled(appName string) bool {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"http-auth:report",
			appName,
		},
	})
	if err != nil {
		return false
	}

	lines := strings.SplitN(result.StdoutContents(), "\n", 2)
	if len(lines) == 0 {
		return false
	}

	parts := strings.SplitN(lines[0], ":", 2)
	if len(parts) < 2 {
		return false
	}

	return strings.TrimSpace(parts[1]) == "true"
}

// enableHttpAuth enables HTTP authentication for an app
func enableHttpAuth(app, username, password string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}
	if httpAuthEnabled(app) {
		state.State = "present"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"http-auth:on",
			app,
			username,
			password,
		},
	})
	state.Commands = append(state.Commands, result.Command)
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "present"
	return state
}

// disableHttpAuth disables HTTP authentication for an app
func disableHttpAuth(app string) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "present",
	}
	if !httpAuthEnabled(app) {
		state.State = "absent"
		return state
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args: []string{
			"--quiet",
			"http-auth:off",
			app,
		},
	})
	state.Commands = append(state.Commands, result.Command)
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = "absent"
	return state
}

// init registers the HttpAuthTask with the task registry
func init() {
	RegisterTask(&HttpAuthTask{})
}
