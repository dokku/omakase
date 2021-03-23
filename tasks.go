package main

type Recipe []struct {
	Inputs []Input         `yaml:"inputs,omitempty"`
	Tasks  []TaskContainer `yaml:"tasks,omitempty"`
}

type Input struct {
	Name        string `yaml:"name"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

type TaskContainer struct {
	Name      string
	DokkuApp  *AppTask  `yaml:"dokku_app,omitempty"`
	DokkuSync *SyncTask `yaml:"dokku_sync,omitempty"`
}

type Task interface {
	DesiredState() string
	NeedsExecution() bool
	Execute() (string, error)
}
