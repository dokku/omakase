#!/usr/bin/env bats

load test_helper

setup() {
    require_dokku
    docket_build
    dokku_clean_app docket-test-plan
}

teardown() {
    dokku_clean_app docket-test-plan
}

@test "docket plan reports drift on a missing app" {
    write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-plan
      dokku_app:
        app: docket-test-plan
EOF
    run "$(docket_bin)" plan --tasks "$TASKS_FILE"
    assert_success
    assert_output --partial "[+]"
    assert_output --partial "Plan:"
    assert_output --partial "1 would change"
}

@test "docket plan does not mutate state" {
    write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-plan
      dokku_app:
        app: docket-test-plan
EOF
    run "$(docket_bin)" plan --tasks "$TASKS_FILE"
    assert_success
    run dokku apps:exists docket-test-plan
    # apps:exists returns non-zero when the app does not exist.
    assert_failure
}

@test "docket plan reports in sync after apply" {
    write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-plan
      dokku_app:
        app: docket-test-plan
EOF
    "$(docket_bin)" apply --tasks "$TASKS_FILE"
    run "$(docket_bin)" plan --tasks "$TASKS_FILE"
    assert_success
    assert_output --partial "[ok]"
    assert_output --partial "in sync"
    assert_output --partial "0 would change"
}

@test "docket plan itemizes config keys to set" {
    "$(docket_bin)" apply --tasks <(cat <<EOF
---
- tasks:
    - dokku_app:
        app: docket-test-plan
EOF
)
    write_tasks_file <<EOF
---
- tasks:
    - name: configure
      dokku_config:
        app: docket-test-plan
        restart: false
        config:
          KEY_ONE: value-one
          KEY_TWO: value-two
EOF
    run "$(docket_bin)" plan --tasks "$TASKS_FILE"
    assert_success
    assert_output --partial "2 key(s) to set"
    assert_output --partial "set KEY_ONE"
    assert_output --partial "set KEY_TWO"
}

@test "docket apply continues to behave as before" {
    write_tasks_file <<EOF
---
- tasks:
    - name: ensure docket-test-plan
      dokku_app:
        app: docket-test-plan
EOF
    run "$(docket_bin)" apply --tasks "$TASKS_FILE"
    assert_success
    run dokku apps:exists docket-test-plan
    assert_success
}
