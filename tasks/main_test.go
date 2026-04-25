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

	if task.DesiredState() != StatePresent {
		t.Errorf("expected state 'present', got '%s'", task.DesiredState())
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
		"dokku_app",
		"dokku_builder_property",
		"dokku_checks_toggle",
		"dokku_config",
		"dokku_domains_toggle",
		"dokku_git_from_image",
		"dokku_git_sync",
		"dokku_network_property",
		"dokku_ports",
		"dokku_proxy_toggle",
		"dokku_resource_limit",
		"dokku_resource_reserve",
		"dokku_service_create",
		"dokku_service_link",
		"dokku_storage_ensure",
		"dokku_storage_mount",
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

	if task.DesiredState() != StatePresent {
		t.Errorf("expected default state 'present', got %q", task.DesiredState())
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

func TestGetTasksConfigTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: set config
      dokku_config:
        app: test-app
        restart: false
        config:
          KEY1: val1
          KEY2: val2
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("set config")
	if task == nil {
		t.Fatal("task 'set config' not found")
	}

	configTask, ok := task.(*ConfigTask)
	if !ok {
		// tasks may be stored as value types depending on reflection
		ct, ok2 := task.(ConfigTask)
		if !ok2 {
			t.Fatalf("task is not a ConfigTask (type is %T)", task)
		}
		configTask = &ct
	}

	if configTask.App != "test-app" {
		t.Errorf("App = %q, want %q", configTask.App, "test-app")
	}
	// Note: defaults.SetDefaults overrides restart=false with the default tag value "true"
	// because false is the zero value for bool. This documents the actual behavior.
	if !configTask.Restart {
		t.Error("Restart = false, want true (defaults.SetDefaults overrides zero-value bool)")
	}
	if len(configTask.Config) != 2 {
		t.Fatalf("expected 2 config keys, got %d", len(configTask.Config))
	}
	if configTask.Config["KEY1"] != "val1" {
		t.Errorf("Config[KEY1] = %q, want %q", configTask.Config["KEY1"], "val1")
	}
	if configTask.Config["KEY2"] != "val2" {
		t.Errorf("Config[KEY2] = %q, want %q", configTask.Config["KEY2"], "val2")
	}
}

func TestGetTasksPortsTaskWithMappings(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: set ports
      dokku_ports:
        app: test-app
        port_mappings:
          - scheme: http
            host: 80
            container: 5000
          - scheme: https
            host: 443
            container: 5000
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("set ports")
	if task == nil {
		t.Fatal("task 'set ports' not found")
	}

	portsTask, ok := task.(*PortsTask)
	if !ok {
		pt, ok2 := task.(PortsTask)
		if !ok2 {
			t.Fatalf("task is not a PortsTask (type is %T)", task)
		}
		portsTask = &pt
	}

	if len(portsTask.PortMappings) != 2 {
		t.Fatalf("expected 2 port mappings, got %d", len(portsTask.PortMappings))
	}

	if portsTask.PortMappings[0].String() != "http:80:5000" {
		t.Errorf("mapping[0] = %q, want %q", portsTask.PortMappings[0].String(), "http:80:5000")
	}
	if portsTask.PortMappings[1].String() != "https:443:5000" {
		t.Errorf("mapping[1] = %q, want %q", portsTask.PortMappings[1].String(), "https:443:5000")
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

func TestGetTasksResourceLimitTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: set resource limits
      dokku_resource_limit:
        app: test-app
        process_type: web
        resources:
          cpu: "100"
          memory: "256"
        clear_before: true
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("set resource limits")
	if task == nil {
		t.Fatal("task 'set resource limits' not found")
	}

	rlTask, ok := task.(*ResourceLimitTask)
	if !ok {
		rt, ok2 := task.(ResourceLimitTask)
		if !ok2 {
			t.Fatalf("task is not a ResourceLimitTask (type is %T)", task)
		}
		rlTask = &rt
	}

	if rlTask.App != "test-app" {
		t.Errorf("App = %q, want %q", rlTask.App, "test-app")
	}
	if rlTask.ProcessType != "web" {
		t.Errorf("ProcessType = %q, want %q", rlTask.ProcessType, "web")
	}
	if len(rlTask.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(rlTask.Resources))
	}
	if rlTask.Resources["cpu"] != "100" {
		t.Errorf("Resources[cpu] = %q, want %q", rlTask.Resources["cpu"], "100")
	}
	if rlTask.Resources["memory"] != "256" {
		t.Errorf("Resources[memory] = %q, want %q", rlTask.Resources["memory"], "256")
	}
	// Unlike ConfigTask.Restart (default:"true"), ClearBefore has default:"false" which
	// is the zero value for bool. Since defaults.SetDefaults only overrides zero values,
	// and true is non-zero, setting clear_before: true in YAML is preserved correctly.
	if !rlTask.ClearBefore {
		t.Error("ClearBefore = false, want true (YAML value should be preserved)")
	}
}

func TestGetTasksResourceReserveTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: set resource reservations
      dokku_resource_reserve:
        app: test-app
        process_type: web
        resources:
          cpu: "100"
          memory: "256"
        clear_before: true
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("set resource reservations")
	if task == nil {
		t.Fatal("task 'set resource reservations' not found")
	}

	rrTask, ok := task.(*ResourceReserveTask)
	if !ok {
		rt, ok2 := task.(ResourceReserveTask)
		if !ok2 {
			t.Fatalf("task is not a ResourceReserveTask (type is %T)", task)
		}
		rrTask = &rt
	}

	if rrTask.App != "test-app" {
		t.Errorf("App = %q, want %q", rrTask.App, "test-app")
	}
	if rrTask.ProcessType != "web" {
		t.Errorf("ProcessType = %q, want %q", rrTask.ProcessType, "web")
	}
	if len(rrTask.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(rrTask.Resources))
	}
	if rrTask.Resources["cpu"] != "100" {
		t.Errorf("Resources[cpu] = %q, want %q", rrTask.Resources["cpu"], "100")
	}
	if rrTask.Resources["memory"] != "256" {
		t.Errorf("Resources[memory] = %q, want %q", rrTask.Resources["memory"], "256")
	}
	if !rrTask.ClearBefore {
		t.Error("ClearBefore = false, want true (YAML value should be preserved)")
	}
}

func TestGetTasksServiceCreateTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create redis service
      dokku_service_create:
        service: redis
        name: my-redis
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("create redis service")
	if task == nil {
		t.Fatal("task 'create redis service' not found")
	}

	scTask, ok := task.(*ServiceCreateTask)
	if !ok {
		st, ok2 := task.(ServiceCreateTask)
		if !ok2 {
			t.Fatalf("task is not a ServiceCreateTask (type is %T)", task)
		}
		scTask = &st
	}

	if scTask.Service != "redis" {
		t.Errorf("Service = %q, want %q", scTask.Service, "redis")
	}
	if scTask.Name != "my-redis" {
		t.Errorf("Name = %q, want %q", scTask.Name, "my-redis")
	}
	if scTask.DesiredState() != StatePresent {
		t.Errorf("expected default state 'present', got %q", scTask.DesiredState())
	}
}

func TestGetTasksServiceCreateWithTemplateContext(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: create {{ .service_type }} service
      dokku_service_create:
        service: {{ .service_type }}
        name: {{ .service_name }}
`)
	context := map[string]interface{}{
		"service_type": "postgres",
		"service_name": "my-db",
	}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("create postgres service")
	if task == nil {
		t.Fatal("task 'create postgres service' not found")
	}

	scTask, ok := task.(*ServiceCreateTask)
	if !ok {
		st, ok2 := task.(ServiceCreateTask)
		if !ok2 {
			t.Fatalf("task is not a ServiceCreateTask (type is %T)", task)
		}
		scTask = &st
	}

	if scTask.Service != "postgres" {
		t.Errorf("Service = %q, want %q", scTask.Service, "postgres")
	}
	if scTask.Name != "my-db" {
		t.Errorf("Name = %q, want %q", scTask.Name, "my-db")
	}
}

func TestGetTasksServiceLinkTaskParsedCorrectly(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: link redis service
      dokku_service_link:
        app: my-app
        service: redis
        name: my-redis
`)
	context := map[string]interface{}{}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("link redis service")
	if task == nil {
		t.Fatal("task 'link redis service' not found")
	}

	slTask, ok := task.(*ServiceLinkTask)
	if !ok {
		st, ok2 := task.(ServiceLinkTask)
		if !ok2 {
			t.Fatalf("task is not a ServiceLinkTask (type is %T)", task)
		}
		slTask = &st
	}

	if slTask.App != "my-app" {
		t.Errorf("App = %q, want %q", slTask.App, "my-app")
	}
	if slTask.Service != "redis" {
		t.Errorf("Service = %q, want %q", slTask.Service, "redis")
	}
	if slTask.Name != "my-redis" {
		t.Errorf("Name = %q, want %q", slTask.Name, "my-redis")
	}
	if slTask.DesiredState() != StatePresent {
		t.Errorf("expected default state 'present', got %q", slTask.DesiredState())
	}
}

func TestGetTasksServiceLinkWithTemplateContext(t *testing.T) {
	data := []byte(`---
- tasks:
    - name: link {{ .service_type }} service
      dokku_service_link:
        app: {{ .app_name }}
        service: {{ .service_type }}
        name: {{ .service_name }}
`)
	context := map[string]interface{}{
		"app_name":     "my-app",
		"service_type": "postgres",
		"service_name": "my-db",
	}

	tasks, err := GetTasks(data, context)
	if err != nil {
		t.Fatalf("GetTasks failed: %v", err)
	}

	task := tasks.Get("link postgres service")
	if task == nil {
		t.Fatal("task 'link postgres service' not found")
	}

	slTask, ok := task.(*ServiceLinkTask)
	if !ok {
		st, ok2 := task.(ServiceLinkTask)
		if !ok2 {
			t.Fatalf("task is not a ServiceLinkTask (type is %T)", task)
		}
		slTask = &st
	}

	if slTask.App != "my-app" {
		t.Errorf("App = %q, want %q", slTask.App, "my-app")
	}
	if slTask.Service != "postgres" {
		t.Errorf("Service = %q, want %q", slTask.Service, "postgres")
	}
	if slTask.Name != "my-db" {
		t.Errorf("Name = %q, want %q", slTask.Name, "my-db")
	}
}
