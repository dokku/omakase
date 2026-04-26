package tasks

import (
	"fmt"
	"github.com/dokku/docket/subprocess"
	"strings"
)

// DomainsTask manages the domains for a given dokku application or globally
type DomainsTask struct {
	// App is the name of the app
	App string `required:"false" yaml:"app"`

	// Global is a flag indicating if the domains should be applied globally
	Global bool `required:"false" yaml:"global,omitempty"`

	// Domains is the list of domain names
	Domains []string `required:"false" yaml:"domains"`

	// State is the desired state of the domains
	State State `required:"false" yaml:"state" default:"present" options:"present,absent,set,clear"`
}

// DomainsTaskExample contains an example of a DomainsTask
type DomainsTaskExample struct {
	// Name is the task name holding the DomainsTask description
	Name string `yaml:"-"`

	// DomainsTask is the DomainsTask configuration
	DomainsTask DomainsTask `yaml:"dokku_domains"`
}

// GetName returns the name of the example
func (e DomainsTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the domains task
func (t DomainsTask) Doc() string {
	return "Manages the domains for a given dokku application or globally"
}

// Examples returns the examples for the domains task
func (t DomainsTask) Examples() ([]Doc, error) {
	return MarshalExamples([]DomainsTaskExample{
		{
			Name: "Add domains to an app",
			DomainsTask: DomainsTask{
				App:     "example-app",
				Domains: []string{"example.com", "www.example.com"},
			},
		},
		{
			Name: "Remove domains from an app",
			DomainsTask: DomainsTask{
				App:     "example-app",
				Domains: []string{"old.example.com"},
				State:   "absent",
			},
		},
		{
			Name: "Set global domains",
			DomainsTask: DomainsTask{
				Global:  true,
				Domains: []string{"global.example.com"},
				State:   "set",
			},
		},
		{
			Name: "Clear all domains from an app",
			DomainsTask: DomainsTask{
				App:   "example-app",
				State: "clear",
			},
		},
	})
}

// Execute manages the domains
func (t DomainsTask) Execute() TaskOutputState {
	return DispatchState(t.State, map[State]func() TaskOutputState{
		StatePresent: func() TaskOutputState { return addDomains(t) },
		StateAbsent:  func() TaskOutputState { return removeDomains(t) },
		StateSet:     func() TaskOutputState { return setDomains(t) },
		StateClear:   func() TaskOutputState { return clearDomains(t) },
	})
}

// validateDomainsTask validates the domains task parameters
func validateDomainsTask(t DomainsTask, requireDomains bool) error {
	if t.Global && t.App != "" {
		return fmt.Errorf("'app' must not be set when 'global' is set to true")
	}
	if !t.Global && t.App == "" {
		return fmt.Errorf("'app' is required when 'global' is not set to true")
	}
	if requireDomains && len(t.Domains) == 0 {
		return fmt.Errorf("'domains' must not be empty for state '%s'", t.State)
	}
	return nil
}

// getDomains fetches current domains for an app or globally
func getDomains(app string, global bool) (map[string]bool, error) {
	reportFlag := "--domains-app-vhosts"
	args := []string{
		"domains:report",
		app,
		reportFlag,
	}
	if global {
		args = []string{
			"domains:report",
			"--global",
			"--domains-global-vhosts",
		}
	}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return nil, err
	}

	domains := map[string]bool{}
	for _, domain := range strings.Fields(result.StdoutContents()) {
		domains[domain] = true
	}
	return domains, nil
}

// addDomains adds domains if they don't already exist
func addDomains(t DomainsTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if err := validateDomainsTask(t, true); err != nil {
		state.Error = err
		return state
	}

	currentDomains, err := getDomains(t.App, t.Global)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	var newDomains []string
	for _, domain := range t.Domains {
		if !currentDomains[domain] {
			newDomains = append(newDomains, domain)
		}
	}

	if len(newDomains) == 0 {
		state.State = StatePresent
		return state
	}

	subcommand := "domains:add"
	appName := t.App
	if t.Global {
		subcommand = "domains:add-global"
		appName = "--global"
	}

	args := []string{"--quiet", subcommand, appName}
	args = append(args, newDomains...)

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

// removeDomains removes domains if they exist
func removeDomains(t DomainsTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StatePresent,
	}

	if err := validateDomainsTask(t, true); err != nil {
		state.Error = err
		return state
	}

	currentDomains, err := getDomains(t.App, t.Global)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	var domainsToRemove []string
	for _, domain := range t.Domains {
		if currentDomains[domain] {
			domainsToRemove = append(domainsToRemove, domain)
		}
	}

	if len(domainsToRemove) == 0 {
		state.State = StateAbsent
		return state
	}

	subcommand := "domains:remove"
	appName := t.App
	if t.Global {
		subcommand = "domains:remove-global"
		appName = "--global"
	}

	args := []string{"--quiet", subcommand, appName}
	args = append(args, domainsToRemove...)

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

// setDomains replaces all domains with the specified ones
func setDomains(t DomainsTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StateAbsent,
	}

	if err := validateDomainsTask(t, true); err != nil {
		state.Error = err
		return state
	}

	subcommand := "domains:set"
	appName := t.App
	if t.Global {
		subcommand = "domains:set-global"
		appName = "--global"
	}

	args := []string{"--quiet", subcommand, appName}
	args = append(args, t.Domains...)

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StateSet
	return state
}

// clearDomains removes all domains
func clearDomains(t DomainsTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   StatePresent,
	}

	if err := validateDomainsTask(t, false); err != nil {
		state.Error = err
		return state
	}

	subcommand := "domains:clear"
	appName := t.App
	if t.Global {
		subcommand = "domains:clear-global"
		appName = "--global"
	}

	args := []string{"--quiet", subcommand, appName}

	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    args,
	})
	if err != nil {
		return TaskOutputErrorFromExec(state, err, result)
	}

	state.Changed = true
	state.State = StateClear
	return state
}

// init registers the DomainsTask with the task registry
func init() {
	RegisterTask(&DomainsTask{})
}
