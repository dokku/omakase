#!/usr/bin/env bash
# Shared helpers for docket bats tests.
#
# Bats and the bats-support / bats-assert helper libraries are expected to be
# installed system-wide via apt or npm; the tests load them from the standard
# /usr/lib/bats/* paths. See .github/workflows/test.yml for the CI install
# steps and tests/bats/README.md for local developer setup.

set -euo pipefail

# Load bats-support / bats-assert from the standard package paths. Both are
# distributed as bash sources we `source` rather than as bats `load` units.
load_bats_libraries() {
    local lib
    for lib in /usr/lib/bats/bats-support/load.bash /usr/lib/bats-support/load.bash; do
        if [ -f "$lib" ]; then
            # shellcheck disable=SC1090
            source "$lib"
            break
        fi
    done
    for lib in /usr/lib/bats/bats-assert/load.bash /usr/lib/bats-assert/load.bash; do
        if [ -f "$lib" ]; then
            # shellcheck disable=SC1090
            source "$lib"
            break
        fi
    done
}

load_bats_libraries

# Resolve the docket binary. Prefer the in-tree build at ./docket so a local
# `go build` tests the working tree; fall back to PATH for environments that
# install the binary system-wide.
docket_bin() {
    if [ -n "${DOCKET_BIN:-}" ]; then
        echo "$DOCKET_BIN"
        return
    fi
    local repo_root
    repo_root="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    if [ -x "$repo_root/docket" ]; then
        echo "$repo_root/docket"
        return
    fi
    command -v docket
}

# docket_build builds the docket binary at the repo root. Subsequent calls in
# the same bats run skip the rebuild because Go caches incremental builds.
docket_build() {
    local repo_root
    repo_root="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    (cd "$repo_root" && go build -o docket .)
    export DOCKET_BIN="$repo_root/docket"
}

# write_tasks_file writes its stdin to "$BATS_TEST_TMPDIR/tasks.yml" and exports
# TASKS_FILE so tests can pass --tasks "$TASKS_FILE".
write_tasks_file() {
    TASKS_FILE="$BATS_TEST_TMPDIR/tasks.yml"
    cat > "$TASKS_FILE"
    export TASKS_FILE
}

# dokku_clean_app destroys an app if it exists. Used in setup/teardown to make
# tests idempotent on shared CI hosts. Ignores missing dokku (developers run
# offline tests too).
dokku_clean_app() {
    local app="$1"
    if ! command -v dokku >/dev/null 2>&1; then
        return 0
    fi
    if dokku apps:exists "$app" >/dev/null 2>&1; then
        dokku --force apps:destroy "$app" >/dev/null 2>&1 || true
    fi
}

# require_dokku skips the current test when no dokku binary is installed,
# matching the integration_helpers_test.go skipIfNoDokkuT helper for Go tests.
require_dokku() {
    if ! command -v dokku >/dev/null 2>&1; then
        skip "dokku not available"
    fi
}
