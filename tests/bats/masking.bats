#!/usr/bin/env bats

load test_helper

setup() {
  require_dokku
  docket_build
  dokku_clean_app docket-test-mask
}

teardown() {
  dokku_clean_app docket-test-mask
}

@test "docket apply --verbose masks an input declared sensitive" {
  write_tasks_file <<'EOF'
---
- inputs:
    - { name: secret_value, required: true, sensitive: true }
  tasks:
    - name: ensure docket-test-mask
      dokku_app:
        app: docket-test-mask
    - name: set the secret
      dokku_config:
        app: docket-test-mask
        config:
          MY_SECRET: "{{ .secret_value }}"
EOF
  run "$(docket_bin)" apply --tasks "$TASKS_FILE" --secret_value=topsecret123 --verbose
  assert_success
  refute_output --partial "topsecret123"
  assert_output --partial "***"
}

@test "docket apply --verbose masks dokku_config map values" {
  write_tasks_file <<'EOF'
---
- tasks:
    - name: ensure docket-test-mask
      dokku_app:
        app: docket-test-mask
    - name: set a literal config value
      dokku_config:
        app: docket-test-mask
        config:
          MY_LITERAL: literal-value-zzz
EOF
  run "$(docket_bin)" apply --tasks "$TASKS_FILE" --verbose
  assert_success
  refute_output --partial "literal-value-zzz"
  # base64 of literal-value-zzz is bGl0ZXJhbC12YWx1ZS16eno=, also masked.
  refute_output --partial "bGl0ZXJhbC12YWx1ZS16eno"
  assert_output --partial "***"
}

@test "DOKKU_TRACE masks values from inputs declared sensitive" {
  write_tasks_file <<'EOF'
---
- inputs:
    - { name: secret_value, required: true, sensitive: true }
  tasks:
    - name: ensure docket-test-mask
      dokku_app:
        app: docket-test-mask
    - name: set the secret
      dokku_config:
        app: docket-test-mask
        config:
          MY_SECRET: "{{ .secret_value }}"
EOF
  DOKKU_TRACE=1 run "$(docket_bin)" apply --tasks "$TASKS_FILE" --secret_value=tracesecretzzz
  assert_success
  refute_output --partial "tracesecretzzz"
}

@test "docket plan output never echoes dokku_config map values" {
  write_tasks_file <<'EOF'
---
- tasks:
    - name: ensure docket-test-mask
      dokku_app:
        app: docket-test-mask
    - name: set a literal config value
      dokku_config:
        app: docket-test-mask
        config:
          MY_LITERAL: literal-value-zzz
EOF
  run "$(docket_bin)" plan --tasks "$TASKS_FILE"
  assert_success
  refute_output --partial "literal-value-zzz"
}
