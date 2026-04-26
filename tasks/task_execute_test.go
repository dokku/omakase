package tasks

import (
	"strings"
	"testing"
)

func TestAppTaskInvalidState(t *testing.T) {
	task := AppTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
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

func TestDomainsTaskInvalidState(t *testing.T) {
	task := DomainsTask{App: "test-app", Domains: []string{"example.com"}, State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestDomainsTaskMissingApp(t *testing.T) {
	task := DomainsTask{Domains: []string{"example.com"}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
}

func TestDomainsTaskGlobalWithApp(t *testing.T) {
	task := DomainsTask{App: "test-app", Global: true, Domains: []string{"example.com"}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestDomainsTaskEmptyDomains(t *testing.T) {
	states := []State{StatePresent, StateAbsent, StateSet}
	for _, s := range states {
		task := DomainsTask{App: "test-app", Domains: []string{}, State: s}
		result := task.Execute()
		if result.Error == nil {
			t.Fatalf("Execute with empty domains and state=%s should return an error", s)
		}
	}
}

func TestDomainsTaskClearNoDomains(t *testing.T) {
	task := DomainsTask{App: "test-app", State: StateClear}
	result := task.Execute()
	// Should fail because dokku isn't running, but NOT because of missing domains
	if result.Error != nil && strings.Contains(result.Error.Error(), "must not be empty") {
		t.Error("clear state should not require domains")
	}
}

func TestDomainsToggleTaskInvalidState(t *testing.T) {
	task := DomainsToggleTask{App: "test-app", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestHttpAuthTaskInvalidState(t *testing.T) {
	task := HttpAuthTask{App: "test-app", Username: "admin", Password: "secret", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestHttpAuthTaskPresentWithoutUsername(t *testing.T) {
	task := HttpAuthTask{App: "test-app", Password: "secret", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no username")
	}
	if !strings.Contains(result.Error.Error(), "username is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestHttpAuthTaskPresentWithoutPassword(t *testing.T) {
	task := HttpAuthTask{App: "test-app", Username: "admin", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no password")
	}
	if !strings.Contains(result.Error.Error(), "password is required") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGitFromImageTaskInvalidState(t *testing.T) {
	task := GitFromImageTask{App: "test-app", Image: "nginx", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestNetworkTaskInvalidState(t *testing.T) {
	task := NetworkTask{Name: "test-network", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestNginxPropertyTaskInvalidState(t *testing.T) {
	task := NginxPropertyTask{App: "test-app", Property: "proxy-read-timeout", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestNginxPropertyTaskMissingApp(t *testing.T) {
	task := NginxPropertyTask{Property: "proxy-read-timeout", Value: "120s", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
}

func TestNginxPropertyTaskGlobalWithAppSet(t *testing.T) {
	task := NginxPropertyTask{
		App:      "test-app",
		Global:   true,
		Property: "proxy-read-timeout",
		Value:    "120s",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestNginxPropertyTaskPresentWithoutValue(t *testing.T) {
	task := NginxPropertyTask{
		App:      "test-app",
		Property: "proxy-read-timeout",
		Value:    "",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no value")
	}
	if !strings.Contains(result.Error.Error(), "invalid without a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestNginxPropertyTaskAbsentWithValue(t *testing.T) {
	task := NginxPropertyTask{
		App:      "test-app",
		Property: "proxy-read-timeout",
		Value:    "120s",
		State:    StateAbsent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when absent state has a value")
	}
	if !strings.Contains(result.Error.Error(), "invalid with a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGitPropertyTaskInvalidState(t *testing.T) {
	task := GitPropertyTask{App: "test-app", Property: "deploy-branch", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestGitPropertyTaskMissingApp(t *testing.T) {
	task := GitPropertyTask{Property: "deploy-branch", Value: "main", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
}

func TestGitPropertyTaskGlobalWithAppSet(t *testing.T) {
	task := GitPropertyTask{
		App:      "test-app",
		Global:   true,
		Property: "deploy-branch",
		Value:    "main",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGitPropertyTaskPresentWithoutValue(t *testing.T) {
	task := GitPropertyTask{
		App:      "test-app",
		Property: "deploy-branch",
		Value:    "",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no value")
	}
	if !strings.Contains(result.Error.Error(), "invalid without a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGitPropertyTaskAbsentWithValue(t *testing.T) {
	task := GitPropertyTask{
		App:      "test-app",
		Property: "deploy-branch",
		Value:    "main",
		State:    StateAbsent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when absent state has a value")
	}
	if !strings.Contains(result.Error.Error(), "invalid with a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestNetworkPropertyTaskInvalidState(t *testing.T) {
	task := NetworkPropertyTask{App: "test-app", Property: "attach-post-create", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestSchedulerPropertyTaskInvalidState(t *testing.T) {
	task := SchedulerPropertyTask{App: "test-app", Property: "selected", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestSchedulerPropertyTaskMissingApp(t *testing.T) {
	task := SchedulerPropertyTask{Property: "selected", Value: "docker-local", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute without app and global=false should return an error")
	}
}

func TestSchedulerPropertyTaskGlobalWithAppSet(t *testing.T) {
	task := SchedulerPropertyTask{
		App:      "test-app",
		Global:   true,
		Property: "selected",
		Value:    "docker-local",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestSchedulerPropertyTaskPresentWithoutValue(t *testing.T) {
	task := SchedulerPropertyTask{
		App:      "test-app",
		Property: "selected",
		Value:    "",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no value")
	}
	if !strings.Contains(result.Error.Error(), "invalid without a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestSchedulerPropertyTaskAbsentWithValue(t *testing.T) {
	task := SchedulerPropertyTask{
		App:      "test-app",
		Property: "selected",
		Value:    "docker-local",
		State:    StateAbsent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when absent state has a value")
	}
	if !strings.Contains(result.Error.Error(), "invalid with a value") {
		t.Errorf("unexpected error: %v", result.Error)
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

func TestPsScaleTaskInvalidState(t *testing.T) {
	task := PsScaleTask{App: "test-app", Scale: map[string]int{"web": 1}, State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestPsScaleTaskEmptyScale(t *testing.T) {
	task := PsScaleTask{App: "test-app", Scale: map[string]int{}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with empty scale and state=present should return an error")
	}
}

func TestPsScaleTaskNilScale(t *testing.T) {
	task := PsScaleTask{App: "test-app", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with nil scale and state=present should return an error")
	}
}

func TestResourceLimitTaskInvalidState(t *testing.T) {
	task := ResourceLimitTask{
		App:       "test-app",
		Resources: map[string]string{"cpu": "100"},
		State:     "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestResourceLimitTaskEmptyResources(t *testing.T) {
	task := ResourceLimitTask{App: "test-app", Resources: map[string]string{}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with empty resources and state=present should return an error")
	}
}

func TestResourceLimitTaskNilResources(t *testing.T) {
	task := ResourceLimitTask{App: "test-app", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with nil resources and state=present should return an error")
	}
}

func TestServiceCreateTaskInvalidState(t *testing.T) {
	task := ServiceCreateTask{Service: "redis", Name: "test-service", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestServiceLinkTaskInvalidState(t *testing.T) {
	task := ServiceLinkTask{App: "test-app", Service: "redis", Name: "test-service", State: "invalid"}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestResourceReserveTaskInvalidState(t *testing.T) {
	task := ResourceReserveTask{
		App:       "test-app",
		Resources: map[string]string{"cpu": "100"},
		State:     "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestResourceReserveTaskEmptyResources(t *testing.T) {
	task := ResourceReserveTask{App: "test-app", Resources: map[string]string{}, State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with empty resources and state=present should return an error")
	}
}

func TestResourceReserveTaskNilResources(t *testing.T) {
	task := ResourceReserveTask{App: "test-app", State: StatePresent}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with nil resources and state=present should return an error")
	}
}

func TestBuilderPropertyTaskGlobalWithAppSet(t *testing.T) {
	task := BuilderPropertyTask{
		App:      "test-app",
		Global:   true,
		Property: "selected",
		Value:    "dockerfile",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestBuilderPropertyTaskPresentWithoutValue(t *testing.T) {
	task := BuilderPropertyTask{
		App:      "test-app",
		Property: "selected",
		Value:    "",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no value")
	}
	if !strings.Contains(result.Error.Error(), "invalid without a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestBuilderPropertyTaskAbsentWithValue(t *testing.T) {
	task := BuilderPropertyTask{
		App:      "test-app",
		Property: "selected",
		Value:    "dockerfile",
		State:    StateAbsent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when absent state has a value")
	}
	if !strings.Contains(result.Error.Error(), "invalid with a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestNetworkPropertyTaskGlobalWithAppSet(t *testing.T) {
	task := NetworkPropertyTask{
		App:      "test-app",
		Global:   true,
		Property: "bind-all-interfaces",
		Value:    "true",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when both global and app are set")
	}
	if !strings.Contains(result.Error.Error(), "must not be set when 'global' is set to true") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestNetworkPropertyTaskPresentWithoutValue(t *testing.T) {
	task := NetworkPropertyTask{
		App:      "test-app",
		Property: "bind-all-interfaces",
		Value:    "",
		State:    StatePresent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when present state has no value")
	}
	if !strings.Contains(result.Error.Error(), "invalid without a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestNetworkPropertyTaskAbsentWithValue(t *testing.T) {
	task := NetworkPropertyTask{
		App:      "test-app",
		Property: "bind-all-interfaces",
		Value:    "true",
		State:    StateAbsent,
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("expected error when absent state has a value")
	}
	if !strings.Contains(result.Error.Error(), "invalid with a value") {
		t.Errorf("unexpected error: %v", result.Error)
	}
}

func TestGitFromImageTaskNonDeployedStates(t *testing.T) {
	tests := []struct {
		name  string
		state State
	}{
		{"present state", StatePresent},
		{"absent state", StateAbsent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := GitFromImageTask{App: "test-app", Image: "nginx", State: tt.state}
			result := task.Execute()
			if result.Error == nil {
				t.Fatal("expected error for non-deployed state")
			}
			if !strings.Contains(result.Error.Error(), "invalid state") {
				t.Errorf("expected 'invalid state' error, got: %v", result.Error)
			}
		})
	}
}

func TestGitSyncTaskInvalidState(t *testing.T) {
	task := GitSyncTask{
		App:    "test-app",
		Remote: "https://example.com/repo",
		State:  "invalid",
	}
	result := task.Execute()
	if result.Error == nil {
		t.Fatal("Execute with invalid state should return an error")
	}
}

func TestPortMappingStringVariousValues(t *testing.T) {
	tests := []struct {
		name string
		pm   PortMapping
		want string
	}{
		{"http standard", PortMapping{Scheme: "http", Host: 80, Container: 5000}, "http:80:5000"},
		{"https", PortMapping{Scheme: "https", Host: 443, Container: 5000}, "https:443:5000"},
		{"high ports", PortMapping{Scheme: "http", Host: 8080, Container: 80}, "http:8080:80"},
		{"zero ports", PortMapping{Scheme: "http", Host: 0, Container: 0}, "http:0:0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pm.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStorageEnsureEmptyChown(t *testing.T) {
	task := StorageEnsureTask{App: "test-app", Chown: "", State: StatePresent}
	result := task.Execute()
	if result.Error == nil || result.Error.Error() != "invalid chown value specified" {
		t.Errorf("expected 'invalid chown value specified' error, got: %v", result.Error)
	}
}

func TestAllTasksExamplesReturnNoError(t *testing.T) {
	for name, task := range RegisteredTasks {
		t.Run(name, func(t *testing.T) {
			_, err := task.Examples()
			if err != nil {
				t.Errorf("Examples() returned error: %v", err)
			}
		})
	}
}

func TestRegisteredTaskCount(t *testing.T) {
	expected := 23
	if got := len(RegisteredTasks); got != expected {
		t.Errorf("expected %d registered tasks, got %d", expected, got)
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
		{&DomainsTask{}, "Manages the domains for a given dokku application or globally"},
		{&DomainsToggleTask{}, "Enables or disables the domains plugin for a given dokku application"},
		{&GitFromImageTask{}, "Deploys a git repository from a docker image"},
		{&GitPropertyTask{}, "Manages the git configuration for a given dokku application"},
		{&GitSyncTask{}, "Syncs a git repository to a dokku application"},
		{&HttpAuthTask{}, "Manages HTTP authentication for a given dokku application"},
		{&NetworkTask{}, "Creates or destroys a Docker network"},
		{&NetworkPropertyTask{}, "Manages the network property for a given dokku application"},
		{&NginxPropertyTask{}, "Manages the nginx configuration for a given dokku application"},
		{&PortsTask{}, "Manages the ports for a given dokku application"},
		{&PsScaleTask{}, "Manages the process scale for a given dokku application"},
		{&ResourceLimitTask{}, "Manages the resource limits for a given dokku application"},
		{&ResourceReserveTask{}, "Manages the resource reservations for a given dokku application"},
		{&SchedulerPropertyTask{}, "Manages the scheduler configuration for a given dokku application"},
		{&ServiceCreateTask{}, "Creates or destroys a dokku service"},
		{&ServiceLinkTask{}, "Links or unlinks a dokku service to an app"},
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
