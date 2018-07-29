# Extending Ingress Monitor Controller

## Adding support for a new Monitor

You can easily implement a new monitor and use it via the controller. First of all you will need to create a folder under `/pkg/monitors/` with the name of the new monitor and then you will create a new service struct inside this folder that implements the following monitor service interface

```go
type MonitorService interface {
    GetAll() []Monitor
    Add(m Monitor)
    Update(m Monitor)
    GetByName(name string) (*Monitor, error)
    Remove(m Monitor)
    Setup(provider Provider)
}
```

*Note:* While developing, make sure to follow the conventions mentioned below in the [Naming Conventions section](#naming-conventions)

Once the implementation of your service is done, you have to open up `monitor-proxy.go` and add a new case inside `OfType` method for your new monitor. Lets say you have named your service `MyNewMonitorService`, then you have to add the case like in the example below:

```go
func (mp *MonitorServiceProxy) OfType(mType string) MonitorServiceProxy {
    mp.monitorType = mType
    switch mType {
    case "UptimeRobot":
        mp.monitor = &UpTimeMonitorService{}
    case "MyNewMonitor":
        mp.monitor = &MyNewMonitorService{}
    default:
        log.Panic("No such provider found")
    }
    return *mp
}
```

Note that the name you specify here for the case will be the key for your new monitor which you can add it in ConfigMap.

Also in case of handling custom api objects for the monitor api, you can create mappers that map from the api objects to the generic `Monitor` objects. The way you have to create these is to create a file named `monitorname-mappers.go` and add mapping functions in that file. An example of a mapping function is found below:

```go
func UptimeMonitorMonitorToBaseMonitorMapper(uptimeMonitor UptimeMonitorMonitor) *Monitor {
    var m Monitor

    m.name = uptimeMonitor.FriendlyName
    m.url = uptimeMonitor.URL
    m.id = strconv.Itoa(uptimeMonitor.ID)

    return &m
}
```

## Naming Conventions

### Annotations

You should use the following format for annotations when there are monitor specific annotations:

```bash
<monitor-name>.monitor.stakater.com/<annotation-name>
```

You should use the following format for annotations when there are global annotations:

```bash
monitor.stakater.com/<annotation-name>
```

#### Examples

For example you're adding support for a new monitor service named `alertme`, it's specific annotations would look like the following:

```bash
alertme.monitor.stakater.com/some-key
```

In case of a global annotation, lets say you want to create 1 for disabling deletion of specific monitors, it would look like so:

```bash
monitor.stakater.com/keep-on-delete
```
