# dokku_cron_property

Manages the cron configuration for a given dokku application

## Setting the mailto address for an app

```yaml
dokku_cron_property:
    app: node-js-app
    property: mailto
    value: ops@example.com
```

## Setting the mailto address globally

```yaml
dokku_cron_property:
    app: ""
    global: true
    property: mailto
    value: ops@example.com
```

## Clearing the mailto address for an app

```yaml
dokku_cron_property:
    app: node-js-app
    property: mailto
```
