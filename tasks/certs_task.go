package tasks

import (
	"fmt"
	"strings"

	"github.com/dokku/docket/subprocess"
)

// CertsTask manages SSL certificates for a dokku app or globally
type CertsTask struct {
	// App is the name of the app. Required if Global is false.
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the certificate should be applied globally
	// via the dokku-global-cert plugin
	Global bool `required:"false" yaml:"global,omitempty"`

	// Cert is the path on the dokku server to the SSL certificate file
	Cert string `required:"false" yaml:"cert,omitempty"`

	// Key is the path on the dokku server to the SSL certificate key file
	Key string `required:"false" yaml:"key,omitempty"`

	// State is the desired state of the SSL configuration
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// CertsTaskExample contains an example of a CertsTask
type CertsTaskExample struct {
	// Name is the task name holding the CertsTask description
	Name string `yaml:"-"`

	// CertsTask is the CertsTask configuration
	CertsTask CertsTask `yaml:"dokku_certs"`
}

// GetName returns the name of the example
func (e CertsTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the certs task
func (t CertsTask) Doc() string {
	return "Manages SSL certificates for a dokku app or globally"
}

// Examples returns the examples for the certs task
func (t CertsTask) Examples() ([]Doc, error) {
	return MarshalExamples([]CertsTaskExample{
		{
			Name: "Add an SSL certificate to an app",
			CertsTask: CertsTask{
				App:  "node-js-app",
				Cert: "/etc/nginx/ssl/node-js-app.crt",
				Key:  "/etc/nginx/ssl/node-js-app.key",
			},
		},
		{
			Name: "Remove an SSL certificate from an app",
			CertsTask: CertsTask{
				App:   "node-js-app",
				State: StateAbsent,
			},
		},
		{
			Name: "Add a global SSL certificate (requires the dokku-global-cert plugin)",
			CertsTask: CertsTask{
				Global: true,
				Cert:   "/etc/nginx/ssl/global.crt",
				Key:    "/etc/nginx/ssl/global.key",
			},
		},
		{
			Name: "Remove the global SSL certificate",
			CertsTask: CertsTask{
				Global: true,
				State:  StateAbsent,
			},
		},
	})
}

// Execute manages the SSL certificate
func (t CertsTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		StatePresent: func() TaskOutputState { return addCert(t) },
		StateAbsent:  func() TaskOutputState { return removeCert(t) },
	})
}

// validateCertsTask validates the certs task parameters
func validateCertsTask(t CertsTask) error {
	if t.Global && t.App != "" {
		return fmt.Errorf("'app' must not be set when 'global' is set to true")
	}
	if !t.Global && t.App == "" {
		return fmt.Errorf("'app' is required when 'global' is not set to true")
	}
	return nil
}

// certsEnabled checks if a certificate is currently configured for an app or globally
func certsEnabled(t CertsTask) (bool, error) {
	args := []string{"--quiet", "certs:report", t.App, "--ssl-enabled"}
	if t.Global {
		args = []string{"--quiet", "global-cert:report", "--global-cert-enabled"}
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(result.StdoutContents()) == "true", nil
}

// addCert installs an SSL certificate for an app or globally
func addCert(t CertsTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if err := validateCertsTask(t); err != nil {
		state.Error = err
		return state
	}
	if t.Cert == "" || t.Key == "" {
		state.Error = fmt.Errorf("'cert' and 'key' are required when state is 'present'")
		return state
	}

	enabled, err := certsEnabled(t)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}
	if enabled {
		state.State = StatePresent
		return state
	}

	args := []string{"--quiet", "certs:add", t.App, t.Cert, t.Key}
	if t.Global {
		args = []string{"--quiet", "global-cert:set", t.Cert, t.Key}
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StatePresent
	return state
}

// removeCert removes the SSL certificate for an app or globally
func removeCert(t CertsTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StatePresent,
	}

	if err := validateCertsTask(t); err != nil {
		state.Error = err
		return state
	}

	enabled, err := certsEnabled(t)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}
	if !enabled {
		state.State = StateAbsent
		return state
	}

	args := []string{"--quiet", "certs:remove", t.App}
	if t.Global {
		args = []string{"--quiet", "global-cert:remove"}
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StateAbsent
	return state
}

// init registers the CertsTask with the task registry
func init() {
	RegisterTask(&CertsTask{})
}
