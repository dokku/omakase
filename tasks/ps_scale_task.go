package tasks

import (
	"fmt"
	"omakase/subprocess"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// PsScaleTask manages the process scale for a given dokku application
type PsScaleTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Scale is a map of process types to quantities
	Scale map[string]int `required:"true" yaml:"scale"`

	// SkipDeploy skips the corresponding deploy
	SkipDeploy bool `yaml:"skip_deploy" default:"false"`

	// State is the desired state of the process scale
	State State `required:"false" yaml:"state,omitempty" default:"present" options:"present"`
}

// PsScaleTaskExample contains an example of a PsScaleTask
type PsScaleTaskExample struct {
	// Name is the task name holding the PsScaleTask description
	Name string `yaml:"-"`

	// PsScaleTask is the PsScaleTask configuration
	PsScaleTask PsScaleTask `yaml:"dokku_ps_scale"`
}

// DesiredState returns the desired state of the process scale
func (t PsScaleTask) DesiredState() State {
	return t.State
}

// Doc returns the docblock for the ps scale task
func (t PsScaleTask) Doc() string {
	return "Manages the process scale for a given dokku application"
}

// Examples returns the examples for the ps scale task
func (t PsScaleTask) Examples() ([]Doc, error) {
	examples := []PsScaleTaskExample{
		{
			Name: "Scale web and worker processes",
			PsScaleTask: PsScaleTask{
				App: "hello-world",
				Scale: map[string]int{
					"web":    2,
					"worker": 1,
				},
			},
		},
		{
			Name: "Scale web and worker processes without deploy",
			PsScaleTask: PsScaleTask{
				App:        "hello-world",
				SkipDeploy: true,
				Scale: map[string]int{
					"web":    4,
					"worker": 4,
				},
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

// Execute sets the process scale for a given dokku application
func (t PsScaleTask) Execute() TaskOutputState {
	if t.State == StatePresent {
		if t.Scale == nil || len(t.Scale) == 0 {
			return TaskOutputState{
				Error: fmt.Errorf("scale must be specified when state is present"),
			}
		}
	}

	funcMap := map[State]func(PsScaleTask) TaskOutputState{
		"present": setPsScale,
	}

	fn, ok := funcMap[t.State]
	if !ok {
		return TaskOutputState{
			Error: fmt.Errorf("invalid state: %s", t.State),
		}
	}
	return fn(t)
}

// getPsScale retrieves the current process scale for a given dokku application
func getPsScale(app string) (map[string]int, error) {
	result, err := subprocess.CallExecCommand(subprocess.ExecCommandInput{
		Command: "dokku",
		Args:    []string{"--quiet", "ps:scale", app},
	})
	if err != nil {
		return nil, err
	}

	scale := map[string]int{}
	for _, line := range strings.Split(result.StdoutContents(), "\n") {
		// strip all whitespace from the line, matching the upstream ansible module
		line = strings.Join(strings.Fields(line), "")
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		qty, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		scale[parts[0]] = qty
	}
	return scale, nil
}

// setPsScale sets the process scale for a given dokku application
func setPsScale(t PsScaleTask) TaskOutputState {
	state := TaskOutputState{
		Changed: false,
		State:   "absent",
	}

	existing, err := getPsScale(t.App)
	if err != nil {
		state.Error = err
		state.Message = err.Error()
		return state
	}

	var proctypesToScale []string
	for proctype, qty := range t.Scale {
		if existingQty, ok := existing[proctype]; ok && existingQty == qty {
			continue
		}
		proctypesToScale = append(proctypesToScale, fmt.Sprintf("%s=%d", proctype, qty))
	}

	if len(proctypesToScale) == 0 {
		state.State = "present"
		return state
	}

	args := []string{
		"ps:scale",
	}

	if t.SkipDeploy {
		args = append(args, "--skip-deploy")
	}

	args = append(args, t.App)
	args = append(args, proctypesToScale...)

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

// init registers the PsScaleTask with the task registry
func init() {
	RegisterTask(&PsScaleTask{})
}
