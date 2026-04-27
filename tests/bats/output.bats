#!/usr/bin/env bats

load test_helper

setup() {
  require_dokku
  docket_build
  dokku_clean_app docket-test-output
}

teardown() {
  dokku_clean_app docket-test-output
}

@test "docket apply prints structured per-task output" {
  write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-output
      dokku_app:
        app: docket-test-output
EOF
  run "$(docket_bin)" apply --tasks "$TASKS_FILE"
  assert_success
  assert_output --partial "==> Play: tasks"
  assert_output --partial "[changed] "
  assert_output --partial "Summary:"
  assert_output --regexp "took [0-9]+\.[0-9]+s"
}

@test "docket apply on second run reports ok" {
  write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-output
      dokku_app:
        app: docket-test-output
EOF
  "$(docket_bin)" apply --tasks "$TASKS_FILE"
  run "$(docket_bin)" apply --tasks "$TASKS_FILE"
  assert_success
  assert_output --partial "[ok]"
  refute_output --partial "[changed]"
  assert_output --partial "0 changed"
}

@test "docket apply --verbose echoes resolved commands" {
  write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-output
      dokku_app:
        app: docket-test-output
EOF
  run "$(docket_bin)" apply --tasks "$TASKS_FILE" --verbose
  assert_success
  assert_output --partial "→ dokku"
  assert_output --partial "apps:create"
}

@test "docket apply on error surfaces the failure" {
  write_tasks_file <<EOF
---
- tasks:
    - name: add ports to a missing app
      dokku_ports:
        app: docket-test-output-missing
        port_mappings:
          - { scheme: http, host_port: 80, container_port: 5000 }
        state: present
EOF
  run "$(docket_bin)" apply --tasks "$TASKS_FILE"
  assert_failure
  assert_output --partial "[error]"
  assert_output --partial "          !"
  assert_output --partial "1 error"
}

@test "docket apply respects NO_COLOR" {
  write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-output
      dokku_app:
        app: docket-test-output
EOF
  NO_COLOR=1 run "$(docket_bin)" apply --tasks "$TASKS_FILE"
  assert_success
  refute_output --partial $'\x1b['
}
