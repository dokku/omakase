# omakase

> Note: this is a heavy work in progress. YMMV.
> Should we change the name? It's a "cute" name but doesn't really help end-users
> and could be seen as fetishisation of Japanese culture. New name suggestions
> welcome!

A method to pre-package and ship applications on Dokku.

## Background

While Ansible is all well and good, having something native to Dokku for shipping applications is awesome. The `omakase` package allows users to specify exactly what it means to be an app, while allowing for some minimal customization (which is against the `omakase` spirit but here we are).

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
omakase
```

A task file can also be specified via flag, and may be a file retrieved via http:

```shell
# alternate path
omakase --tasks path/to/task.yml

# html file
omakase --tasks http://dokku.com/omakase/example.yml
```

Some other ideas:

- This could be automatically applied from within a repository if a .dokku/tasks.yml was found. In such a case, certain tasks would be added to a denylist and would be ignored during the run (such as dokku_app or dokku_sync).
- Dokku may expose a command such as dokku app:install that would allow users to invoke omakase to install apps.
- A web ui could expose a web ui to customize remote task files and then call `omakase` directly on the generated output.

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
omakase --name lollipop
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

If all inputs are specified on the CLI, then they are injected as is. Otherwise, unless the `--no-interactive` flag is specified, `omakase` will ask for values for each input, with the cli-specified values merged onto the task file default values as defaults.

Finally, the following input keys are reserved for internal usage:

- `help`
- `tasks`
- `v`
- `version`

### Tasks

All implemented tasks should closely follow those available via the `ansible-dokku` library. Additionally, `omakase` will expose a few custom tasks that are specific to this package to ease migration from pure ansible.

Tasks will have both a `name` and an execution context, where the context maps to a single implemented modules. Tasks can be templated out via the variables from the `inputs` section, and may also use any functions exposed by `gliderlabs/sigil`.

#### Adding a new task

Task executors should be added by creating an `execute_TASK_NAME.go`. The Task name should be `lower_underscore_case`. By way of example, an `example_lollipop.go` would contain the following:

```go
package main

type LollipopTask struct {
  App   string `required:"true" yaml:"app"`
  State string `required:"true" yaml:"state" default:"present"`
}

func (t LollipopTask) DesiredState() string {
  return t.State
}

func (t LollipopTask) NeedsExecution() bool {
  return true
}

func (t LollipopTask) Execute() (string, error) {
  return "", nil
}

func (t *LollipopTask) SetDefaultDesiredState(state string) {
    if t.State == "" {
        t.State = state
    }
}
```

The `LollipopTask` struct contains the fields necessary for the task. The only necessary field is `State`, which holds the desired state of the task. All other fields are completely custom for the task at hand.

The `DesiredState()` function must return `t.State`.

The `Execute()` function should actually execute the task. The return values:

- `string`: a string holding the current state
- `error`: Whether an error occurred during processing

> Todo: How do we expose stdout? Should the Status object actually be more structured? Should it serialize to json directly for use by ansible?

The `NeedsExecution()` function should check if the task should be executed. This may always return true, or may execute something to figure out if the task still needs execution.

The `SetDefaultDesiredState()` function can be implemented as shown above, and is used internally to ensure that the task has a default state if the user did not specify one.