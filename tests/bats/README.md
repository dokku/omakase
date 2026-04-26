# bats tests

End-to-end tests for the `docket` CLI, exercised against a real Dokku
installation. These complement the Go unit tests under `tasks/*_test.go`
(which mock subprocess) and the Go integration tests under
`tasks/*_integration_test.go` (which exercise the task layer against a real
Dokku).

## Layout

- `test_helper.bash` - shared helpers for building the binary, writing
  per-test `tasks.yml` fixtures, and cleaning up apps in setup/teardown.
- `*.bats` - one file per CLI subcommand.

## Running locally

The tests need:

- `bats-core` (`bats` on PATH)
- `bats-support` and `bats-assert` (loaded from `/usr/lib/bats/*` or
  `/usr/lib/bats-support/`, `/usr/lib/bats-assert/`)
- A Dokku installation reachable via the `dokku` CLI

On Debian / Ubuntu:

```bash
sudo apt-get install -y bats bats-support bats-assert
```

Or via npm:

```bash
sudo npm install -g bats
# bats-support / bats-assert have no npm distribution; install via apt or from source
```

Run the suite from the repo root:

```bash
bats tests/bats/
```

Tests skip themselves when `dokku` is not available, so the suite is safe to
run on a developer laptop without a local Dokku.

## CI

`.github/workflows/test.yml` defines a `bats-test` job that installs Dokku
and the bats helper packages, builds the docket binary, and runs the suite.
