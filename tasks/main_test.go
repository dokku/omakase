package tasks

import (
	"strings"
	"testing"

	_ "github.com/gliderlabs/sigil/builtin"
)

func TestGetTasksEmptyRecipe(t *testing.T) {
	data := []byte("---\n")
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("GetTasks with empty recipe should return an error")
	}

	if !strings.Contains(err.Error(), "no recipe found") {
		t.Errorf("expected 'no recipe found' error, got: %v", err)
	}
}

func TestGetTasksEmptyList(t *testing.T) {
	data := []byte("---\n- tasks: []\n")
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks with empty task list should not error, got: %v", err)
	}

	if len(tasks.Keys()) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks.Keys()))
	}
}

func TestGetTasksValidAppTask(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create test app
      dokku_app:
        app: test-app
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks should not error for valid app task, got: %v", err)
	}

	if len(tasks.Keys()) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks.Keys()))
	}

	task := tasks.Get("create test app")
	if task == nil {
		t.Fatal("task 'create test app' not found")
	}

	appTask, ok := task.(*AppTask)
	if !ok {
		t.Fatalf("task is not an AppTask (type is %T)", task)
	}
	if appTask.State != StatePresent {
		t.Errorf("expected state 'present', got '%s'", appTask.State)
	}
}

func TestGetTasksInvalidTaskType(t *testing.T) {
	data := []byte(`---
- tasks:
    - dokku_nonexistent:
        app: test-app
`)
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("GetTasks with invalid task type should return an error")
	}

	if !strings.Contains(err.Error(), "not a valid task") {
		t.Errorf("expected 'not a valid task' error, got: %v", err)
	}
}

func TestGetTasksTooManyProperties(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: test
      dokku_app:
        app: test-app
      dokku_config:
        app: test-app
`)
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("GetTasks with too many properties should return an error")
	}

	if !strings.Contains(err.Error(), "too many properties") {
		t.Errorf("expected 'too many properties' error, got: %v", err)
	}
}

func TestGetTasksAutoGeneratesName(t *testing.T) {
	data := []byte(`---
- tasks:
    - dokku_app:
        app: test-app
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks should not error for nameless task, got: %v", err)
	}

	keys := tasks.Keys()
	if len(keys) != 1 {
		t.Fatalf("expected 1 task, got %d", len(keys))
	}

	if !strings.HasPrefix(keys[0], "task #1 ") {
		t.Errorf("expected auto-generated name starting with 'task #1 ', got '%s'", keys[0])
	}
}

func TestGetTasksWithTemplateContext(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create {{ .app_name }}
      dokku_app:
        app: {{ .app_name }}
`)
	context := map[string]interface{}{
		"app_name": "my-app",
	}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks should not error with template context, got: %v", err)
	}

	if len(tasks.Keys()) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks.Keys()))
	}

	task := tasks.Get("create my-app")
	if task == nil {
		t.Fatal("task 'create my-app' not found")
	}
}

func TestRegisteredTasksExist(t *testing.T) {
	expectedTasks := []string{
		"dokku_acl_app",
		"dokku_acl_service",
		"dokku_app",
		"dokku_app_clone",
		"dokku_app_json_property",
		"dokku_app_lock",
		"dokku_builder_dockerfile_property",
		"dokku_builder_herokuish_property",
		"dokku_builder_lambda_property",
		"dokku_builder_nixpacks_property",
		"dokku_builder_pack_property",
		"dokku_builder_property",
		"dokku_builder_railpack_property",
		"dokku_buildpacks",
		"dokku_caddy_property",
		"dokku_certs",
		"dokku_checks_property",
		"dokku_checks_toggle",
		"dokku_config",
		"dokku_cron_property",
		"dokku_docker_options",
		"dokku_domains",
		"dokku_domains_toggle",
		"dokku_git_auth",
		"dokku_git_from_archive",
		"dokku_git_from_image",
		"dokku_git_property",
		"dokku_git_sync",
		"dokku_haproxy_property",
		"dokku_http_auth",
		"dokku_letsencrypt",
		"dokku_letsencrypt_property",
		"dokku_logs_property",
		"dokku_network",
		"dokku_network_property",
		"dokku_nginx_property",
		"dokku_openresty_property",
		"dokku_ports",
		"dokku_proxy_toggle",
		"dokku_ps_property",
		"dokku_ps_scale",
		"dokku_registry_auth",
		"dokku_registry_property",
		"dokku_resource_limit",
		"dokku_resource_reserve",
		"dokku_scheduler_docker_local_property",
		"dokku_scheduler_k3s_property",
		"dokku_scheduler_property",
		"dokku_service_create",
		"dokku_service_link",
		"dokku_storage_ensure",
		"dokku_storage_mount",
		"dokku_traefik_property",
	}

	for _, name := range expectedTasks {
		if _, ok := RegisteredTasks[name]; !ok {
			t.Errorf("expected task %q to be registered", name)
		}
	}
}

func TestGetTasksMultipleTasks(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create app
      dokku_app:
        app: test-app
    - name: set config
      dokku_config:
        app: test-app
        config:
          KEY: VALUE
    - name: mount storage
      dokku_storage_mount:
        app: test-app
        host_dir: /host
        container_dir: /container
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	keys := tasks.Keys()
	if len(keys) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(keys))
	}

	expectedNames := []string{"create app", "set config", "mount storage"}
	for i, name := range expectedNames {
		if keys[i] != name {
			t.Errorf("task[%d] = %q, want %q", i, keys[i], name)
		}
		if tasks.Get(name) == nil {
			t.Errorf("task %q not found", name)
		}
	}
}

func TestGetTasksTaskWithDefaultState(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create app
      dokku_app:
        app: test-app
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("create app")
	if task == nil {
		t.Fatal("task not found")
	}

	appTask, ok := task.(*AppTask)
	if !ok {
		t.Fatalf("task is not an AppTask (type is %T)", task)
	}
	if appTask.State != StatePresent {
		t.Errorf("expected default state 'present', got %q", appTask.State)
	}
}

func TestGetTasksInvalidYaml(t *testing.T) {
	data := []byte("not valid yaml: [[[")
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestGetTasksSigilTemplateError(t *testing.T) {
	data := []byte(`---
- tasks:
    - dokku_app:
        app: {{ .broken
`)
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("expected error for bad template syntax")
	}
	if !strings.Contains(err.Error(), "re-render error") {
		t.Errorf("expected 're-render error', got: %v", err)
	}
}

func TestGetTasksTwoPropertiesNoName(t *testing.T) {
	data := []byte(`---
- tasks:
    - dokku_app:
        app: test-app
      dokku_config:
        app: test-app
`)
	context := map[string]interface{}{}

	_, err := GetTasks(data, context)
	if err == nil {
		t.Fatal("expected error for two properties without name")
	}
	if !strings.Contains(err.Error(), "unexpected property") {
		t.Errorf("expected 'unexpected property' error, got: %v", err)
	}
}

func TestGetTasksFromRealExample(t *testing.T) {
	data := []byte(`---
- inputs:
    - name: app
      description: "Name of app to be deployed"
      type: string
      required: true
    - name: image
      default: "lscr.io/linuxserver/adguardhome-sync:latest"
      description: "Image to be deployed"
  tasks:
    - name: create app
      dokku_app:
        app: {{ .app | default "" }}
    - name: set config
      dokku_config:
        app: {{ .app | default "" }}
        restart: false
        config:
          PUID: "1000"
          PGID: "1000"
          TZ: "Europe/UTC"
          CONFIGFILE: "/config/adguardhome-sync.yaml"
    - name: ensure storage
      dokku_storage_ensure:
        app: {{ .app | default "" }}
        chown: "heroku"
    - name: mount storage
      dokku_storage_mount:
        app: {{ .app | default "" }}
        host_dir: "/var/lib/dokku/data/storage/{{ .app | default "" }}"
        container_dir: "/config"
    - name: set ports
      dokku_ports:
        app: {{ .app | default "" }}
        port_mappings:
          - scheme: http
            host: 80
            container: 8080
    - name: deploy image
      dokku_git_from_image:
        app: {{ .app | default "" }}
        image: {{ .image | default "" }}
`)
	context := map[string]interface{}{
		"app":   "test-adguard",
		"image": "lscr.io/linuxserver/adguardhome-sync:latest",
	}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	keys := tasks.Keys()
	if len(keys) != 6 {
		t.Fatalf("expected 6 tasks, got %d", len(keys))
	}

	expectedNames := []string{
		"create app",
		"set config",
		"ensure storage",
		"mount storage",
		"set ports",
		"deploy image",
	}
	for i, name := range expectedNames {
		if keys[i] != name {
			t.Errorf("task[%d] = %q, want %q", i, keys[i], name)
		}
	}

	// verify template context was applied to app task
	appTaskRaw := tasks.Get("create app")
	appTask, ok := appTaskRaw.(*AppTask)
	if !ok {
		at, ok2 := appTaskRaw.(AppTask)
		if !ok2 {
			t.Fatalf("create app is not an AppTask (type is %T)", appTaskRaw)
		}
		appTask = &at
	}
	if appTask.App != "test-adguard" {
		t.Errorf("AppTask.App = %q, want %q", appTask.App, "test-adguard")
	}

	// verify config task has expected config keys
	configTaskRaw := tasks.Get("set config")
	configTask2, ok := configTaskRaw.(*ConfigTask)
	if !ok {
		ct, ok2 := configTaskRaw.(ConfigTask)
		if !ok2 {
			t.Fatalf("set config is not a ConfigTask (type is %T)", configTaskRaw)
		}
		configTask2 = &ct
	}
	if len(configTask2.Config) != 4 {
		t.Errorf("expected 4 config keys, got %d", len(configTask2.Config))
	}

	// verify port mapping was parsed
	portsTaskRaw := tasks.Get("set ports")
	portsTask2, ok := portsTaskRaw.(*PortsTask)
	if !ok {
		pt, ok2 := portsTaskRaw.(PortsTask)
		if !ok2 {
			t.Fatalf("set ports is not a PortsTask (type is %T)", portsTaskRaw)
		}
		portsTask2 = &pt
	}
	if len(portsTask2.PortMappings) != 1 {
		t.Errorf("expected 1 port mapping, got %d", len(portsTask2.PortMappings))
	}
	if portsTask2.PortMappings[0].String() != "http:80:8080" {
		t.Errorf("port mapping = %q, want %q", portsTask2.PortMappings[0].String(), "http:80:8080")
	}

	// verify git from image was parsed with template context
	gitTaskRaw := tasks.Get("deploy image")
	gitTask, ok := gitTaskRaw.(*GitFromImageTask)
	if !ok {
		gt, ok2 := gitTaskRaw.(GitFromImageTask)
		if !ok2 {
			t.Fatalf("deploy image is not a GitFromImageTask (type is %T)", gitTaskRaw)
		}
		gitTask = &gt
	}
	if gitTask.Image != "lscr.io/linuxserver/adguardhome-sync:latest" {
		t.Errorf("Image = %q, want %q", gitTask.Image, "lscr.io/linuxserver/adguardhome-sync:latest")
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
	expected := 54
	if got := len(RegisteredTasks); got != expected {
		t.Errorf("expected %d registered tasks, got %d", expected, got)
	}
}

func TestTaskDocStrings(t *testing.T) {
	tests := []struct {
		task Task
		want string
	}{
		{&AclAppTask{}, "Manages the dokku-acl access list for a dokku application"},
		{&AclServiceTask{}, "Manages the dokku-acl access list for a dokku service"},
		{&AppTask{}, "Creates or destroys an app"},
		{&AppCloneTask{}, "Clones an existing dokku app to a new app"},
		{&AppJsonPropertyTask{}, "Manages the app.json configuration for a given dokku application"},
		{&AppLockTask{}, "Locks or unlocks a dokku application from deployment"},
		{&BuilderDockerfilePropertyTask{}, "Manages the builder-dockerfile configuration for a given dokku application"},
		{&BuilderHerokuishPropertyTask{}, "Manages the builder-herokuish configuration for a given dokku application"},
		{&BuilderLambdaPropertyTask{}, "Manages the builder-lambda configuration for a given dokku application"},
		{&BuilderNixpacksPropertyTask{}, "Manages the builder-nixpacks configuration for a given dokku application"},
		{&BuilderPackPropertyTask{}, "Manages the builder-pack configuration for a given dokku application"},
		{&BuilderPropertyTask{}, "Manages the builder configuration for a given dokku application"},
		{&BuilderRailpackPropertyTask{}, "Manages the builder-railpack configuration for a given dokku application"},
		{&BuildpacksPropertyTask{}, "Manages the buildpacks configuration for a given dokku application"},
		{&BuildpacksTask{}, "Manages the buildpacks for a given dokku application"},
		{&CaddyPropertyTask{}, "Manages the caddy configuration for a given dokku application"},
		{&CertsTask{}, "Manages SSL certificates for a dokku app or globally. The `cert` and `key` fields are paths on the dokku server, so when running with `DOKKU_HOST` set the referenced files must already exist on the remote host - docket does not upload them."},
		{&ChecksPropertyTask{}, "Manages the checks configuration for a given dokku application"},
		{&ChecksToggleTask{}, "Enables or disables the checks plugin for a given dokku application"},
		{&ConfigTask{}, "Manages the configuration for a given dokku application"},
		{&CronPropertyTask{}, "Manages the cron configuration for a given dokku application"},
		{&DockerOptionsTask{}, "Manages docker-options for a given dokku application"},
		{&DomainsTask{}, "Manages the domains for a given dokku application or globally"},
		{&DomainsToggleTask{}, "Enables or disables the domains plugin for a given dokku application"},
		{&GitAuthTask{}, "Manages netrc credentials for a git host"},
		{&GitFromArchiveTask{}, "Deploys a git repository from an archive URL"},
		{&GitFromImageTask{}, "Deploys a git repository from a docker image"},
		{&GitPropertyTask{}, "Manages the git configuration for a given dokku application"},
		{&GitSyncTask{}, "Syncs a git repository to a dokku application"},
		{&HaproxyPropertyTask{}, "Manages the haproxy configuration for a given dokku application"},
		{&HttpAuthTask{}, "Manages HTTP authentication for a given dokku application"},
		{&LetsencryptTask{}, "Enables or disables letsencrypt SSL certificates for a dokku application"},
		{&LetsencryptPropertyTask{}, "Manages the letsencrypt configuration for a given dokku application"},
		{&LogsPropertyTask{}, "Manages the logs configuration for a given dokku application"},
		{&NetworkTask{}, "Creates or destroys a Docker network"},
		{&NetworkPropertyTask{}, "Manages the network property for a given dokku application"},
		{&NginxPropertyTask{}, "Manages the nginx configuration for a given dokku application"},
		{&OpenrestyPropertyTask{}, "Manages the openresty configuration for a given dokku application"},
		{&PortsTask{}, "Manages the ports for a given dokku application"},
		{&PsScaleTask{}, "Manages the process scale for a given dokku application"},
		{&RegistryAuthTask{}, "Manages docker registry authentication for a dokku application or globally"},
		{&RegistryPropertyTask{}, "Manages the registry configuration for a given dokku application"},
		{&ResourceLimitTask{}, "Manages the resource limits for a given dokku application"},
		{&ResourceReserveTask{}, "Manages the resource reservations for a given dokku application"},
		{&SchedulerDockerLocalPropertyTask{}, "Manages the scheduler-docker-local configuration for a given dokku application"},
		{&SchedulerK3sPropertyTask{}, "Manages the scheduler-k3s configuration for a given dokku application"},
		{&SchedulerPropertyTask{}, "Manages the scheduler configuration for a given dokku application"},
		{&ServiceCreateTask{}, "Creates or destroys a dokku service"},
		{&ServiceLinkTask{}, "Links or unlinks a dokku service to an app"},
		{&ProxyToggleTask{}, "Enables or disables the proxy plugin for a given dokku application"},
		{&StorageEnsureTask{}, "Ensures the storage for a given dokku application"},
		{&StorageMountTask{}, "Mounts or unmounts the storage for a given dokku application"},
		{&TraefikPropertyTask{}, "Manages the traefik configuration for a given dokku application"},
	}

	for _, tt := range tests {
		doc := tt.task.Doc()
		if doc != tt.want {
			t.Errorf("Doc() = %q, want %q", doc, tt.want)
		}
	}
}
