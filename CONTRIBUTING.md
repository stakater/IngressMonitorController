# Contributing

## Workflow

Pull Requests are welcome. In general, we follow the "fork-and-pull" Git workflow.

1. **Fork** the repo on GitHub
2. **Clone** the project to your own machine
3. **Commit** changes to your own branch
4. **Push** your work back up to your fork
5. Submit a **Pull request** so that we can review your changes

NOTE: Be sure to merge the latest from "upstream" before making a pull request!

## Golang code practice

Follow this [tour](https://tour.golang.org/) to practice golang.

## Extending Ingress Monitor Controller

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
    Equal(oldMonitor Monitor, newMonitor Monitor) bool
}
```

_Note:_ While developing, make sure to follow the conventions mentioned below in the [Naming Conventions section](#naming-conventions)

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
        panic("No such provider found")
    }
    return *mp
}
```

Similarly, add a new case for your provider in ExtractConfig:

```go
func (mp *MonitorServiceProxy) ExtractConfig(spec endpointmonitorv1alpha1.EndpointMonitorSpec) interface{} {
	var config interface{}

	switch mp.monitorType {
	case "UptimeRobot":
		config = spec.UptimeRobotConfig
    case "MyNewMonitor":
        config = spec.MyNewMonitorConfig
	default:
		return config
	}
	return config
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

### Naming Conventions

#### Configuration

You can add additional configuration in [endpointmonitor_types.go](./pkg/apis/endpointmonitor/v1alpha1/endpointmonitor_types.go)
And then handle it accordingly in your monitor's implementation.

#### Examples

In [endpointmonitor_types.go](./pkg/apis/endpointmonitor/v1alpha1/endpointmonitor_types.go)

```yaml
// UptimeRobotConfig defines the configuration for UptimeRobot Monitor Provider
type UptimeRobotConfig struct {
// The uptimerobot alertContacts to be associated with this monitor
// +optional
AlertContacts string `json:"alertContacts,omitempty"`

// The uptimerobot check interval in seconds
// +kubebuilder:validation:Minimum=60
// +optional
Interval int `json:"interval,omitempty"`

// Specify maintenanceWindows i.e. once or recurring “do-not-monitor periods”
// +optional
MaintenanceWindows string `json:"maintenanceWindows,omitempty"`

// The uptimerobot monitor type (http or keyword)
// +kubebuilder:validation:Enum=http;keyword
// +optional
MonitorType string `json:"monitorType,omitempty"`

// Alert if value exist (yes) or doesn't exist (no) (Only if monitor-type is keyword)
// +kubebuilder:validation:Enum=yes;no
// +optional
KeywordExists string `json:"keywordExists,omitempty"`

// keyword to check on URL (e.g.'search' or '404') (Only if monitor-type is keyword)
// +optional
KeywordValue string `json:"keywordValue,omitempty"`

// The uptimerobot public status page ID to add this monitor to
// +optional
StatusPages string `json:"statusPages,omitempty"`
}
```

And then handle this configuration as handled in `processProviderConfig` in [uptime-monitor.go](./pkg/monitors/uptimerobot/uptime-monitor.go)

## Development

### Dependencies

1. GoLang v1.18
2. kubectl
3. operator-sdk v1.6.2

### Running Operator Locally

1. Install CRDs by running `make install`
2. Create a namespace `test`
3. Create a secret with name `imc-config` and add your desired config in there
4. Run `OPERATOR_NAMESPACE=test make run`

**NOTE**: Ensure that all required resources are re-generated

## Testing

### Running Tests Locally

Tests require a Kubernetes instance to talk to with a `test` namespace created, and a config with a valid UptimeRobot `apiKey` and `alertContacts`. For example, on MacOS with Homebrew and Minikube, you could accomplish this like

```bash
# while still in the root folder, configure test setup
$ export CONFIG_FILE_PATH=$(pwd)/examples/configs/test-config.yaml
# update the apikey and alertContacts in this file and the config_test.go file (`correctTestAPIKey` and `correctTestAlertContacts` contstants)
$ kubectl create namespace test

# run the following command in the root folder
$ make test
```

### Test config for monitors

When running monitor test cases, make sure to provide a config similar to the following:

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
  - name: Pingdom
    apiToken: <your-api-token>
    apiURL: https://api.pingdom.com
    alertIntegrations: "91166-10924"
    alertContacts: "1234567_8_9-9876543_2_1,1234567_8_9-9876543_2_2"
    teamAlertContacts: "1234567_8_9-9876543_2_1,1234567_8_9-9876543_2_2"
enableMonitorDeletion: true
monitorNameTemplate: "{{.Name}}-{{.Namespace}}"
```
