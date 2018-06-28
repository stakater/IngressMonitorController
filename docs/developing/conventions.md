# Naming Conventions

## Annotations

You should use the following format for annotations when there are monitor specific annotations:

```bash
<monitor-name>.monitor.stakater.com/<annotation-name>
```

You should use the following format for annotations when there are global annotations:

```bash
monitor.stakater.com/<annotation-name>
```

### Examples

For example you're adding support for a new monitor service named `alertme`, it's specific annotations would look like the following:

```bash
alertme.monitor.stakater.com/some-key
```

In case of a global annotation, lets say you want to create 1 for disabling deletion of specific monitors, it would look like so:

```bash
monitor.stakater.com/keep-on-delete
```
