# ![](assets/web/IMC-round-100px.png) Ingress Monitor Controller

[![Get started with Stakater](https://stakater.github.io/README/stakater-github-banner.png)](http://stakater.com/?utm_source=IngressMonitorController&utm_medium=github)

## Problem Statement

How do I get notified if any of my services is down?

We want to get notified in a slack channel & email if any of our services become unhealthy!

We want to monitor ingresses in a kubernetes cluster via any uptime checker but the problem is to manually check for new ingresses / removed ingresses and add them to the checker or remove them. There isn't any out of the box solution for this as of now.

## Solution

This controller will continuously watch ingresses in all namespaces, and automatically add / remove monitors in any of the uptime checkers. With the help of this solution, you can keep a check on your services and see whether they're up and running and live.

## Supported Uptime Checkers

Currently we support the following monitors:

- [UptimeRobot](https://uptimerobot.com)
- [Pingdom](https://pingdom.com) ([Additional Config](https://github.com/stakater/IngressMonitorController/blob/master/docs/pingdom-configuration.md))
- [StatusCake](https://www.statuscake.com) ([Additional Config](https://github.com/stakater/IngressMonitorController/blob/master/docs/statuscake-configuration.md))

## Deploying to Kubernetes

### Supported Kubernetes versions

Ingress Monitor Controller has been tested with Kubernetes version 1.8.x, and should work with higher versions.

### Vanilla Manifests

You have to first clone or download the repository contents. The kubernetes deployment and files are provided inside `kubernetes/manifests` folder.

#### Enabling

By default, the controller ignores the ingresses without a specific annotation on it. You will need to add the following annotation on your ingresses so that the controller is able to recognize and monitor the ingresses.

```yaml
"monitor.stakater.com/enabled": "true"
```

The annotation key is `monitor.stakater.com/enabled` and you can toggle the value of this annotation between `true` and `false` to enable or disable monitoring of that specific ingress.

#### Configuring

First of all you need to modify `configmap.yaml`'s `config.yaml` file. Following are the available options that you can use to customize the controller:

| Key                   |Description                                                                    |
|-----------------------|-------------------------------------------------------------------------------|
| providers             | An array of uptime providers that you want to add to your controller          |
| enableMonitorDeletion | A safeguard flag that is used to enable or disable monitor deletion on ingress deletion (Useful for prod environments where you don't want to remove monitor on ingress deletion) |
| watchNamespace        | Name of the namespace if you want to monitor ingresses only in that namespace. Defaults to null |

For the list of providers, there's a number of options that you need to specify. The table below lists them:

| Key           | Description                                                               |
|---------------|---------------------------------------------------------------------------|
| name          | Name of the provider (From the list of supported uptime checkers)         |
| apiKey        | ApiKey of the provider                                                    |
| apiURL        | Base url of the ApiProvider                                               |
| alertContacts | A `-` separated list of contact id's that you want to add to the monitors |

*Note:* Follow [this](https://github.com/stakater/IngressMonitorController/blob/master/docs/fetching-alert-contacts-from-uptime-robot.md) guide to see how to fetch `alertContacts` from UpTimeRobot

#### Configuring through ingress annotations

The following optional annotations allow you to set further configuration:

| Annotation                            | Description                                                                                                                 |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| `"monitor.stakater.com/forceHttps"`   | If set to the string `"true"`, the monitor endpoint will use HTTPS, even if the Ingress manifest itself doesn't specify TLS |
| `"monitor.stakater.com/overridePath"` | Set this annotation to define the healthcheck path for this monitor, overriding the controller's default logic              |
| `"monitor.stakater.com/name"`         | Set this annotation to define the Monitor friendly name in Uptime Robot. If unset, defaults to the template in the config   |

#### Deploying

You can deploy the controller in the namespace you want to monitor by running the following kubectl commands:

```bash
kubectl apply -f configmap.yaml -n <namespace>
kubectl apply -f rbac.yaml -n <namespace>
kubectl apply -f deployment.yaml -n <namespace>
```

*Note*: Before applying rbac.yaml, You need to modify the namespace in the `RoleBinding` subjects section to the namespace you want to apply rbac.yaml to.

#### Running Ingress Monitor Controller for a single namespace

By default the controller watches in all namespaces. If you want to use Ingress Monitor Controller for a specific namespace in your cluster, you can follow 1 of the steps below:

##### Via Helm Chart

Set `watchNamespace` to `<namespace-name>` in `values.yaml` before applying the helm chart and the controller will then watch in that namespace.

##### Via Kubernetes Manifests

Add environment variable `KUBERNETES_NAMESPACE` in `deployments/kubernetes/manifests/deployment.yaml` and set its value to the namespace you want to watch in. After that, apply the manifests located under `deployments/kubernetes/examples/watch-all-namespaces/` in any namespace. The deployed controller will now watch only that namespace.

### Helm Charts

If you configured `helm` on your cluster, you can deploy the controller via helm chart located under `kubernetes/chart` folder.

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

## Help

**Got a question?**
File a GitHub [issue](https://github.com/stakater/IngressMonitorController/issues), or send us an [email](mailto:stakater@gmail.com).

### Talk to us on Slack
Join and talk to us on the #tools-imc channel for discussing the Ingress Monitor Controller

[![Join Slack](https://stakater.github.io/README/stakater-join-slack-btn.png)](https://stakater-slack.herokuapp.com/)
[![Chat](https://stakater.github.io/README/stakater-chat-btn.png)](https://stakater.slack.com/messages/CA66MMYSE/)

## Contributing

### Bug Reports & Feature Requests

Please use the [issue tracker](https://github.com/stakater/IngressMonitorController/issues) to report any bugs or file feature requests.

### Developing

PRs are welcome. In general, we follow the "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that we can review your changes

NOTE: Be sure to merge the latest from "upstream" before making a pull request!

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
enableMonitorDeletion: true
monitorNameTemplate: "{{.IngressName}}-{{.Namespace}}"
```

For example if you want to run only test cases for `StatusCake`, the 1st block of provider should still be present since test cases are written in that way.

## Changelog

View our closed [Pull Requests](https://github.com/stakater/IngressMonitorController/pulls?q=is%3Apr+is%3Aclosed).

## License

Apache2 © [Stakater](http://stakater.com)

## About

The `IngressMonitorController` is maintained by [Stakater][website]. Like it? Please let us know at <hello@stakater.com>

See [our other projects][community]
or contact us in case of professional services and queries on <hello@stakater.com>

  [website]: http://stakater.com/
  [community]: https://www.stakater.com/projects-overview.html

## Contributers

Stakater Team and the Open Source community! :trophy:
