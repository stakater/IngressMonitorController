
### Developing

PRs are welcome. In general, we follow the "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that we can review your changes

NOTE: Be sure to merge the latest from "upstream" before making a pull request!

### Adding support for a new Monitor

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

*Note:* While developing, make sure to follow the conventions mentioned [here](https://github.com/stakater/IngressMonitorController/blob/master/docs/developing/conventions.md)

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

### Running Tests Locally

Tests require a Kubernetes instance to talk to with a `test` namespace created, and a config with a valid UptimeRobot `apiKey` and `alertContacts`. For example, on MacOS with Homebrew and Minikube, you could accomplish this like

```bash
# install dependencies
$ brew install glide
$ glide update

# while still in the root folder, configure test setup
$ export CONFIG_FILE_PATH=$(pwd)/configs/testConfigs/test-config.yaml
# update the apikey and alertContacts in this file and the config_test.go file (`correctTestAPIKey` and `correctTestAlertContacts` contstants)
$ minikube start
$ kubectl create namespace test

# run the following command in the root folder
$ make test
```

### Test config for monitors

When running monitor test cases, make sure to provide a config similar to the following, making sure that the order of providers is the same as below:

```yaml
providers:
  - name: UptimeRobot
    apiKey: <your-api-key>
    apiURL: https://api.uptimerobot.com/v2/
    alertContacts: <your-alert-contacts>
  - name: StatusCake
    apiKey: <your-api-key>
    apiURL: https://app.statuscake.com/API/
    username: <your-account-username>
    password: <your-account-password>
enableMonitorDeletion: true
monitorNameTemplate: "{{.IngressName}}-{{.Namespace}}"
```

For example if you want to run only test cases for `StatusCake`, the 1st block of provider should still be present since test cases are written in that way.
