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

Running `docket` with no subcommand prints the available commands. Use `docket apply` to execute a task file, `docket plan` to preview the changes a task file would make without mutating any state, or `docket version` to print the binary's version.

### Previewing changes with `plan`

`docket plan` reads each task's current state from the live dokku server and reports what `apply` would change, without invoking any mutating dokku command. Each task line is prefixed with a marker:

| Marker | Meaning |
|--------|---------|
| `[ok]` | Task is in sync; `apply` would not change anything |
| `[+]` | `apply` would create new state |
| `[~]` | `apply` would modify existing state |
| `[-]` | `apply` would remove existing state |
| `[error]` | The read-state probe itself errored (drift unknown) |

Tasks that perform multiple operations (e.g. `dokku_config` setting several keys) report each individual mutation under the task line:

```text
[~]       configure  (2 key(s) to set)
          - set KEY_ONE (new)
          - set KEY_TWO (was set)

Plan: 1 task(s); 1 would change, 0 in sync, 0 error(s).
```

`Plan()` results drive `apply`: every task probes the server once, and `apply` reuses that probe to decide whether to mutate. `apply` on an already-converged server reports `Changed=false` for every task; back-to-back applies are no-ops by design.

A handful of tasks (notably `dokku_git_auth`, `dokku_registry_auth`, and `dokku_storage_ensure`) cannot probe their current state without invoking the corresponding dokku command, so their plan output reports drift unconditionally with `(... not probed)` in the reason.

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
