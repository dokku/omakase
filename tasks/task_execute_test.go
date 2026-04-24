package tasks

import (
	"testing"
)

func TestAppTaskInvalidState(t *testing.T) {
	task := AppTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestAppTaskDesiredState(t *testing.T) {
	task := AppTask{App: "test-app", State: StatePresent}
	if task.DesiredState() != StatePresent {
		t.Errorf("expected state 'present', got '%s'", task.DesiredState())
	}

	task = AppTask{App: "test-app", State: StateAbsent}
	if task.DesiredState() != StateAbsent {
		t.Errorf("expected state 'absent', got '%s'", task.DesiredState())
	}
}

func TestBuilderPropertyTaskInvalidState(t *testing.T) {
	task := BuilderPropertyTask{App: "test-app", Property: "selected", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestBuilderPropertyTaskMissingApp(t *testing.T) {
	task := BuilderPropertyTask{Property: "selected", Value: "dockerfile", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
}

func TestChecksToggleTaskInvalidState(t *testing.T) {
	task := ChecksToggleTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestConfigTaskInvalidState(t *testing.T) {
	task := ConfigTask{
		App:    "test-app",
		Config: map[string]string{"KEY": "VALUE"},
		State:  "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestDomainsToggleTaskInvalidState(t *testing.T) {
	task := DomainsToggleTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestGitFromImageTaskInvalidState(t *testing.T) {
	task := GitFromImageTask{App: "test-app", Image: "nginx", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestNetworkPropertyTaskInvalidState(t *testing.T) {
	task := NetworkPropertyTask{App: "test-app", Property: "attach-post-create", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestNetworkPropertyTaskMissingApp(t *testing.T) {
	task := NetworkPropertyTask{Property: "attach-post-create", Value: "test-network", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
}

func TestPortsTaskInvalidState(t *testing.T) {
	task := PortsTask{
		App:          "test-app",
		PortMappings: []PortMapping{{Scheme: "http", Host: 80, Container: 5000}},
		State:        "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestPortsTaskEmptyPortMappings(t *testing.T) {
	task := PortsTask{App: "test-app", PortMappings: []PortMapping{}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with empty port mappings should return an error")
	}
}

func TestProxyToggleTaskInvalidState(t *testing.T) {
	task := ProxyToggleTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestStorageEnsureTaskInvalidState(t *testing.T) {
	task := StorageEnsureTask{App: "test-app", Chown: "heroku", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestStorageMountTaskInvalidState(t *testing.T) {
	task := StorageMountTask{
		App:          "test-app",
		HostDir:      "/host",
		ContainerDir: "/container",
		State:        "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestPortMappingString(t *testing.T) {
	pm := PortMapping{Scheme: "http", Host: 80, Container: 5000}
	expected := "http:80:5000"
	if pm.String() != expected {
		t.Errorf("PortMapping.String() = %q, want %q", pm.String(), expected)
	}
}

func TestStorageEnsureValidChownValues(t *testing.T) {
	validValues := []string{"heroku", "herokuish", "paketo", "root", "false"}
	for _, chown := range validValues {
		task := StorageEnsureTask{App: "test-app", Chown: chown, State: StatePresent}
		result := task.Execute()
		// These will fail because dokku isn't running, but should NOT fail
		// due to invalid chown value
		if result.Error != nil && result.Error.Error() == "invalid chown value specified" {
			t.Errorf("chown value %q should be valid but was rejected", chown)
		}
	}
}

func TestStorageEnsureInvalidChownValue(t *testing.T) {
	task := StorageEnsureTask{App: "test-app", Chown: "packeto", State: StatePresent}
	result := task.Execute()
	if result.Error == nil || result.Error.Error() != "invalid chown value specified" {
		t.Errorf("chown value 'packeto' (misspelled) should be rejected as invalid")
	}
}

func TestStorageEnsureAbsentStateReturnsError(t *testing.T) {
	task := StorageEnsureTask{App: "test-app", Chown: "heroku", State: StateAbsent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with absent state should return an error for storage ensure")
	}
}

func TestGitSyncTaskDesiredState(t *testing.T) {
	task := GitSyncTask{
		App:        "test-app",
		Repository: "https://github.com/example/repo",
		State:      "synced",
	}
	if task.DesiredState() != "synced" {
		t.Errorf("expected state 'synced', got '%s'", task.DesiredState())
	}
}

func TestTaskDocStrings(t *testing.T) {
	tests := []struct {
		task Task
		want string
	}{
		{&AppTask{}, "Creates or destroys an app"},
		{&BuilderPropertyTask{}, "Manages the builder configuration for a given dokku application"},
		{&ChecksToggleTask{}, "Enables or disables the checks plugin for a given dokku application"},
		{&ConfigTask{}, "Manages the configuration for a given dokku application"},
		{&DomainsToggleTask{}, "Enables or disables the domains plugin for a given dokku application"},
		{&GitFromImageTask{}, "Deploys a git repository from a docker image"},
		{&GitSyncTask{}, "Syncs a git repository to a dokku application"},
		{&NetworkPropertyTask{}, "Manages the network property for a given dokku application"},
		{&PortsTask{}, "Manages the ports for a given dokku application"},
		{&ProxyToggleTask{}, "Enables or disables the proxy plugin for a given dokku application"},
		{&StorageEnsureTask{}, "Ensures the storage for a given dokku application"},
		{&StorageMountTask{}, "Mounts or unmounts the storage for a given dokku application"},
	}

	for _, tt := range tests {
		doc := tt.task.Doc()
		if doc != tt.want {
			t.Errorf("Doc() = %q, want %q", doc, tt.want)
		}
	}
}
