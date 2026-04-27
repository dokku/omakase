# docket

> Note: this is a heavy work in progress. YMMV.

A method to pre-package and ship applications on Dokku.

## Background

While Ansible is all well and good, having something native to Dokku for shipping applications is awesome. The `docket` package allows users to specify exactly what it means to be an app, while allowing for some minimal customization.

This package provides the above functionality by exposing the modules from `ansible-dokku` within a single Golang binary. Users of `ansible-dokku` based task lists should be able to use their existing tasks with minimal changes, while organizations can decide to expose apps in easy to use methods for their users.

## Building

```shell
go build
```

## Usage

Create a `tasks.yml` file:

```yaml
---
- tasks:
    - dokku_app:
        app: inflector
    - dokku_sync:
        app: inflector
        repository: http://github.com/cakephp/inflector.cakephp.org
```

Run it:

```shell
# from the same directory as the tasks.yml
docket apply
```

Running `docket` with no subcommand prints the available commands. Use `docket init` to scaffold a starter `tasks.yml`, `docket apply` to execute a task file, `docket fmt` to canonically format a task file, `docket plan` to preview the changes a task file would make without mutating any state, `docket validate` to check a task file's schema and templates without contacting the server, or `docket version` to print the binary's version.

### Scaffolding with `init`

`docket init` writes a starter `tasks.yml` from an embedded template. It is offline only: no Dokku server contact, no `git` subprocess. The default scaffold ships four tasks (`dokku_app`, `dokku_config`, `dokku_domains`, `dokku_git_sync`) wrapped in a single play with `app` and `repo` inputs, and round-trips cleanly through `docket validate`.

```shell
# Use cwd basename as the app and remote.origin.url from ./.git/config as the repo
docket init

# Stream the rendered YAML to stdout for piping
docket init --output -
```

The flags are:

| Flag | Effect |
|------|--------|
| (default) | Write `./tasks.yml`; refuse if the file exists |
| `--output <path>` | Write to a specific path; `-` writes to stdout |
| `--force` | Overwrite an existing file |
| `--name <name>` | Override the play and `app` input default (defaults to the cwd basename) |
| `--repo <url>` | Override the `repo` input default (defaults to `remote.origin.url` in `./.git/config`, if present) |
| `--minimal` | One-task example with no `inputs:` block |

### Formatting recipes with `fmt`

`docket fmt` is a canonical formatter for `tasks.yml`, in the spirit of `gofmt`. It parses with `gopkg.in/yaml.v3`'s `Node` API so head, line, and foot comments round-trip; reorders task envelope and play keys into a stable order; normalises indentation to a 2-space step; and inserts blank lines between top-level plays and between top-level task entries. The default rewrites `./tasks.yml` in place; `--check` and `--diff` are read-only modes. The CLI flags compose, modeled after `black` / `ruff format`.

```shell
# Rewrite ./tasks.yml in place.
docket fmt

# CI gate: print the diff and exit 1 if anything is not canonical.
docket fmt --check --diff

# Read from stdin, write canonical to stdout.
cat tasks.yml | docket fmt -
```

The flags are:

| Flag | Effect |
|------|--------|
| (default) | Format `./tasks.yml` in place; no-op (mtime preserved) when already canonical |
| `--check` | Exit 1 if any file is not canonical; no writes. Composes with `--diff` |
| `--diff` | Print a GNU unified diff against canonical; no writes. Composes with `--check` |
| `--color <when>` | When to colorize the diff: `auto` (default; on if stdout is a TTY and `NO_COLOR` is unset), `always`, `never` |
| `-` | Read from stdin, write canonical to stdout |
| `<path...>` | Format the named files; each argument is expanded as a glob and rewritten in place |

The diff output is GNU unified diff with `--- <path>` / `+++ <path>` / `@@` headers and is consumable by `git apply` and `patch -p0` once colors are stripped.

Before writing, `fmt` re-parses its canonical output and aborts if the round-tripped AST does not match the input AST - a guard against `yaml.v3` emitter edge cases (notably anchors and complex flow scalars). On a parse error or round-trip mismatch the file is not touched and `fmt` exits 1.

### Applying recipes with `apply`

`docket apply` runs every task in the recipe, mutating the live dokku server as needed. Each task line is prefixed with a status marker padded to a fixed column:

| Marker | Meaning |
|--------|---------|
| `[ok]` | Task ran, no change |
| `[changed]` | Task ran, mutated state |
| `[skipped]` | Task was filtered out (tags, `when:`, `--start-at-task`) |
| `[error]` | Task errored; the run aborts |

A play header (`==> Play: tasks`) precedes the per-task lines, and an end-of-run summary line follows them:

```text
==> Play: tasks
[changed] dokku apps:create api
[ok]      dokku config:set api KEY=value

Summary: 2 tasks · 1 changed · 1 ok · 0 skipped · 0 errors  (took 0.8s)
```

On error, the failing task's error message is printed as a `!`-prefixed continuation line and the run aborts with exit 1. The summary still prints with the partial counts before exit.

The flags are:

| Flag | Effect |
|------|--------|
| `--tasks <path>` | Use a specific task file (default `./tasks.yml`) |
| `--verbose` | After each task line, echo the resolved Dokku command on a `→`-prefixed continuation line. Commands are not masked - avoid on recipes that pass secrets via task arguments. |

Color output respects [`NO_COLOR`](https://no-color.org/): set `NO_COLOR=1` to disable ANSI escapes, or pipe to a non-TTY (output is plain in that case automatically).

### Previewing changes with `plan`

`docket plan` reads each task's current state from the live dokku server and reports what `apply` would change, without invoking any mutating dokku command. The output uses the same play header and column layout as `apply`, with a different marker set:

| Marker | Meaning |
|--------|---------|
| `[ok]` | Task is in sync; `apply` would not change anything |
| `[+]` | `apply` would create new state |
| `[~]` | `apply` would modify existing state |
| `[-]` | `apply` would remove existing state |
| `[!]` | The read-state probe itself errored (drift unknown) |

Tasks that perform multiple operations (e.g. `dokku_config` setting several keys) report each individual mutation under the task line:

```text
==> Play: tasks
[~]       configure  (2 key(s) to set)
          - set KEY_ONE (new)
          - set KEY_TWO (was set)

Plan: 1 task(s); 1 would change, 0 in sync, 0 error(s).
```

`Plan()` results drive `apply`: every task probes the server once, and `apply` reuses that probe to decide whether to mutate. `apply` on an already-converged server reports `Changed=false` for every task; back-to-back applies are no-ops by design.

A handful of tasks (notably `dokku_git_auth`, `dokku_registry_auth`, and `dokku_storage_ensure`) cannot probe their current state without invoking the corresponding dokku command, so their plan output reports drift unconditionally with `(... not probed)` in the reason.

### Validating recipes with `validate`

`docket validate` performs offline schema and template checks against a `tasks.yml` without contacting any Dokku server, suitable for CI lint jobs that need to reject broken recipes before deploy.

The shipping checks cover: YAML parses, recipe shape (top-level list of plays with `inputs`/`tasks`), task entry shape (envelope keys plus exactly one task-type key), task type registered (with a "did you mean" suggestion for typos), required fields decode, and sigil templates render against input defaults.

```shell
docket validate --tasks path/to/tasks.yml
```

Exit codes are `0` when no problems are found and `1` otherwise. Two flags are available:

- `--json` emits one JSON-lines event per problem with a stable `version: 1` schema (`{"type":"validate_problem","code":"unknown_task_type", ...}`), suitable for piping into a CI annotator.
- `--strict` additionally flags any input declared `required: true` that has no `default` and no value supplied via a CLI flag - useful in CI to ensure the recipe can be applied without runtime overrides.

A task file can also be specified via flag, and may be a file retrieved via http:

```shell
# alternate path
docket apply --tasks path/to/task.yml

# html file
docket apply --tasks http://dokku.com/docket/example.yml
```

Some other ideas:

- This could be automatically applied from within a repository if a `.dokku/task.yml` was found. In such a case, certain tasks would be added to a denylist and would be ignored during the run (such as dokku_app or dokku_sync).
- Dokku may expose a command such as dokku app:install that would allow users to invoke docket to install apps.
- A web ui could expose a web ui to customize remote task files and then call `docket` directly on the generated output.

### Inputs

Each app recipe can have custom inputs as specified in the `tasks.yml`. Inputs should _not_ reference any variable context, and are extracted using a two-phase parsing method (extract-then-inject).

```yaml
---
- inputs:
    - name: name
      default: "inflector"
      description: "Name of app to be created"
      required: true
  tasks:
    - dokku_app:
        app: {{ .name }}
    - dokku_sync:
        app: {{ .name }}
        repository: http://github.com/cakephp/inflector.cakephp.org
```

With the above, the following method is used to override the `name` variable. Omitting will use the default value.

```shell
# from the same directory as the tasks.yml
docket apply --name lollipop
```

Any inputs for a given task file will also show up in the `--help` output.

Inputs are injected using golang's `text/template` package via the `gliderlabs/sigil` library, and as such have access to everything `gliderlabs/sigil` does.

Inputs can have the following properties:

- name:
  - type: `string`
  - default: ``
- default:
  - type: `bool|float|int|string`
  - default: zero-value for the type
- description:
  - type: `string`
  - default: `""`
- required:
  - type: `bool`
  - default: `false`
- type:
  - type: string
  - default `string`
  - options:
    - `bool`
    - `float`
    - `int`
    - `string`

If all inputs are specified on the CLI, then they are injected as is. Otherwise, unless the `--no-interactive` flag is specified, `docket` will ask for values for each input, with the cli-specified values merged onto the task file default values as defaults.

Finally, the following input keys are reserved for internal usage:

- `help`
- `tasks`
- `v`
- `version`

### Tasks

All implemented tasks should closely follow those available via the `ansible-dokku` library. Additionally, `docket` will expose a few custom tasks that are specific to this package to ease migration from pure ansible.

Tasks will have both a `name` and an execution context, where the context maps to a single implemented modules. Tasks can be templated out via the variables from the `inputs` section, and may also use any functions exposed by `gliderlabs/sigil`.

#### Adding a new task

Task executors should be added by creating an `tasks/${TASK_NAME}_task.go`. The Task name should be `lower_underscore_case`. By way of example, a `tasks/lollipop_task.go` would contain the following:

```go
package main

type LollipopTask struct {
  App   string `required:"true" yaml:"app"`
  State State `required:"true" yaml:"state" default:"present"`
}

func (t LollipopTask) Plan() PlanResult {
  return DispatchPlan(t.State, map[State]func() PlanResult{
    "present": func() PlanResult {
      // Probe the server once, decide whether to mutate.
      if /* already in desired state */ {
        return PlanResult{InSync: true, Status: PlanStatusOK}
      }
      return PlanResult{
        InSync:    false,
        Status:    PlanStatusCreate, // or PlanStatusModify, PlanStatusDestroy
        Reason:    "...",
        Mutations: []string{"create lollipop"},
        apply: func() TaskOutputState {
          // Run the underlying dokku command. Return Changed=true on success.
          return TaskOutputState{Changed: true, State: StatePresent}
        },
      }
    },
    "absent": func() PlanResult { /* ... */ },
  })
}

func (t LollipopTask) Execute() TaskOutputState {
  return ExecutePlan(t.Plan())
}

func init() {
    RegisterTask(&LollipopTask{})
}
```

The `LollipopTask` struct contains the fields necessary for the task. The only necessary field is `State`, which holds the desired state of the task. All other fields are completely custom for the task at hand.

`Plan()` is the canonical implementation: it probes the live server once, computes the diff, and returns a `PlanResult`. When `InSync` is `false`, `Plan()` embeds an `apply` closure that performs the underlying mutation. For tasks that perform multiple operations (e.g. setting several config keys in one call), populate `PlanResult.Mutations` with one entry per atomic change so the plan output can itemize the diff.

`Execute()` is always `return ExecutePlan(t.Plan())`. The shared `ExecutePlan` helper handles the InSync, error, and apply cases uniformly so the per-state mutation logic lives in exactly one place per task.

`DispatchPlan` and `DispatchState` automatically set `DesiredState` on the returned result.

The `init()` function registers the task for usage within a recipe.
