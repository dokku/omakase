# dokku_builder_property

Manages the builder configuration for a given dokku application

## Overriding the auto-selected builder

```yaml
builder_property:
    app: node-js-app
    property: selected
    value: dockerfile
```

## Setting the builder to the default value

```yaml
builder_property:
    app: node-js-app
    property: selected
```

## Changing the build build directory

```yaml
builder_property:
    app: monorepo
    property: build-dir
    value: backend
```

## Overriding the auto-selected builder globally

```yaml
builder_property:
    app: ""
    global: true
    property: selected
    value: herokuish
```
