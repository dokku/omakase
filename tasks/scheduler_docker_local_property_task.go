package tasks

// SchedulerDockerLocalPropertyTask manages the scheduler-docker-local configuration for a given dokku application
type SchedulerDockerLocalPropertyTask struct {
	// App is the name of the app
	App string `required:"true" yaml:"app"`

	// Property is the name of the scheduler-docker-local property to set
	Property string `required:"true" yaml:"property"`

	// Value is the value to set for the scheduler-docker-local property
	Value string `required:"false" yaml:"value,omitempty"`

	// State is the desired state of the scheduler-docker-local configuration
	State State `required:"true" yaml:"state,omitempty" default:"present" options:"present,absent"`
}

// SchedulerDockerLocalPropertyTaskExample contains an example of a SchedulerDockerLocalPropertyTask
type SchedulerDockerLocalPropertyTaskExample struct {
	// Name is the task name holding the SchedulerDockerLocalPropertyTask description
	Name string `yaml:"-"`

	// SchedulerDockerLocalPropertyTask is the SchedulerDockerLocalPropertyTask configuration
	SchedulerDockerLocalPropertyTask SchedulerDockerLocalPropertyTask `yaml:"dokku_scheduler_docker_local_property"`
}

// GetName returns the name of the example
func (e SchedulerDockerLocalPropertyTaskExample) GetName() string {
	return e.Name
}

// Doc returns the docblock for the scheduler-docker-local property task
func (t SchedulerDockerLocalPropertyTask) Doc() string {
	return "Manages the scheduler-docker-local configuration for a given dokku application"
}

// Examples returns the examples for the scheduler-docker-local property task
func (t SchedulerDockerLocalPropertyTask) Examples() ([]Doc, error) {
	return MarshalExamples([]SchedulerDockerLocalPropertyTaskExample{
		{
			Name: "Enabling the init process for an app",
			SchedulerDockerLocalPropertyTask: SchedulerDockerLocalPropertyTask{
				App:      "node-js-app",
				Property: "init-process",
				Value:    "true",
			},
		},
		{
			Name: "Setting the parallel schedule count for an app",
			SchedulerDockerLocalPropertyTask: SchedulerDockerLocalPropertyTask{
				App:      "node-js-app",
				Property: "parallel-schedule-count",
				Value:    "4",
			},
		},
		{
			Name: "Clearing the init process for an app",
			SchedulerDockerLocalPropertyTask: SchedulerDockerLocalPropertyTask{
				App:      "node-js-app",
				Property: "init-process",
			},
		},
	})
}

// Execute sets or unsets the scheduler-docker-local property
func (t SchedulerDockerLocalPropertyTask) Execute() TaskOutputState {
	return ExecutePlan(t.Plan())
}

// Plan reports the drift the SchedulerDockerLocalPropertyTask would produce.
func (t SchedulerDockerLocalPropertyTask) Plan() PlanResult {
	return planProperty(t.State, t.App, false, t.Property, t.Value, "scheduler-docker-local:set")
}

// init registers the SchedulerDockerLocalPropertyTask with the task registry
func init() {
	RegisterTask(&SchedulerDockerLocalPropertyTask{})
}
