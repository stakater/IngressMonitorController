# ![](assets/web/IMC-round-100px.png) Ingress Monitor Controller

[![Get started with Stakater](https://stakater.github.io/README/stakater-github-banner.png)](http://stakater.com/?utm_source=IngressMonitorController&utm_medium=github)

## Problem Statement

How do I get notified if any of my services is down?

We want to get notified in a slack channel & email if any of our services become unhealthy!

We want to monitor ingresses in a kubernetes cluster via any uptime checker but the problem is to manually check for new ingresses / removed ingresses and add them to the checker or remove them. There isn't any out of the box solution for this as of now.

## Solution

This controller will continuously watch ingresses in specific or all namespaces, and automatically add / remove monitors in any of the uptime checkers. With the help of this solution, you can keep a check on your services and see whether they're up and running and live.

## Supported Uptime Checkers

Currently we support the following monitors:

- [UptimeRobot](https://uptimerobot.com)
- [Pingdom](https://pingdom.com) ([Additional Config](https://github.com/stakater/IngressMonitorController/blob/master/docs/pingdom-configuration.md))
- [StatusCake](https://www.statuscake.com) ([Additional Config](https://github.com/stakater/IngressMonitorController/blob/master/docs/statuscake-configuration.md))

## Deploying to Kubernetes

Ingress Monitor Controller can be used to watch ingresses in a specific namespace, or all namespaces. By default the
 controller watches in all namespaces. To use Ingress Monitor Controller for a specific namespace, the relevant
 instructions where indicated should be followed.  

### Supported Kubernetes versions

Ingress Monitor Controller has been tested with Kubernetes version 1.8.x and 1.10.x, and should work with higher versions.

### Enabling

By default, the controller ignores the ingresses without a specific annotation on it. You will need to add the following annotation on your ingresses so that the controller is able to recognize and monitor the ingresses.

```yaml
"monitor.stakater.com/enabled": "true"
```

The annotation key is `monitor.stakater.com/enabled` and you can toggle the value of this annotation between `true` and `false` to enable or disable monitoring of that specific ingress.

### Configuration

Following are the available options that you can use to customize the controller:

| Key                   |Description                                                                    |
|-----------------------|-------------------------------------------------------------------------------|
| providers             | An array of uptime providers that you want to add to your controller          |
| enableMonitorDeletion | A safeguard flag that is used to enable or disable monitor deletion on ingress deletion (Useful for prod environments where you don't want to remove monitor on ingress deletion) |
| watchNamespace        | Name of the namespace if you want to monitor ingresses only in that namespace. Defaults to null |

For the list of providers, there are a number of options that you need to specify. The table below lists them:

| Key           | Description                                                               |
|---------------|---------------------------------------------------------------------------|
| name          | Name of the provider (From the list of supported uptime checkers)         |
| apiKey        | ApiKey of the provider                                                    |
| apiURL        | Base url of the ApiProvider                                               |
| alertContacts | A `-` separated list of contact id's that you want to add to the monitors |

*Note:* Follow [this](https://github.com/stakater/IngressMonitorController/blob/master/docs/fetching-alert-contacts-from-uptime-robot.md) guide to see how to fetch `alertContacts` from UpTimeRobot.
For other uptime checkers refer to the corresponding configuration document from [here](docs/) for any additional
 configuration needed for the specific uptime checker.

#### Configuring through ingress annotations

The following optional annotations allow you to set further configuration:

| Annotation                            | Description                                                                                                                 |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------|
| `"monitor.stakater.com/forceHttps"`   | If set to the string `"true"`, the monitor endpoint will use HTTPS, even if the Ingress manifest itself doesn't specify TLS |
| `"monitor.stakater.com/overridePath"` | Set this annotation to define the healthcheck path for this monitor, overriding the controller's default logic              |
| `"monitor.stakater.com/name"`         | Set this annotation to define the Monitor friendly name in Uptime Robot. If unset, defaults to the template in the config   |

### Vanilla Manifests

The Ingress Monitor Controller can be deployed with vanilla manifests or Helm Charts. For Vanilla manifests, download the
 [ingressmonitorcontroller.yaml](https://github.com/stakater/IngressMonitorController/blob/master/deployments/kubernetes/ingressmonitorcontroller.yaml) file.

#### Configuring

The configuration discussed in the above section needs to be done by modifying `config.yaml` data for the ConfigMap resource in the `ingressmonitorcontroller.yaml` file. 

##### Running for a single namespace

Add environment variable `KUBERNETES_NAMESPACE` in `ingressmonitorcontroller.yaml` for the Deployment resource and set its value
 to the namespace you want to watch in. After that, apply the `ingressmonitorcontroller.yaml` manifest in any namespace.
  The deployed controller will now watch only that namespace.

#### Deploying

You can deploy the controller in the namespace you want to monitor by running the following kubectl command:

```bash
kubectl apply -f ingressmonitorcontroller.yaml -n <namespace>
```

*Note*: Before applying `ingressmonitorcontroller.yaml`, You need to modify the namespace in the `RoleBinding` subjects section to the namespace you want to apply RBAC to.

### Helm Charts

The Ingress Monitor Controller can be deployed with Helm Charts or vanilla manifests. For Helm Charts follow the below steps:

1. Add the chart repo:

   i. `helm repo add stakater https://stakater.github.io/stakater-charts/`

   ii. `helm repo update`
2. Set configuration as discussed in the Configuration section

   i.`helm fetch --untar stakater/ingressmonitorcontroller`

   ii. Open and edit `ingressmonitorcontroller/values.yaml` in a text editor
3. Install the chart
   * `helm install stakater/ingressmonitorcontroller -f ingressmonitorcontroller/values.yaml -n ingressmonitorcontroller`

##### Running for a single namespace

Set `watchNamespace` to `<namespace-name>` in `deployments/kubernetes/chart/ingressmonitorcontroller/values.yaml` before applying the helm chart and the controller will then watch in that namespace.

#### Deploying

You can deploy the controller by running the following command:

```bash
helm install ./deployments/kubernetes/chart/ingressmonitorcontroller --name ingressmonitorcontroller
```

## Help

**Got a question?**
File a GitHub [issue](https://github.com/stakater/IngressMonitorController/issues), or send us an [email](mailto:hello@stakater.com).

### Talk to us on Slack
Join and talk to us on the #tools-imc channel for discussing the Ingress Monitor Controller

[![Join Slack](https://stakater.github.io/README/stakater-join-slack-btn.png)](https://stakater-slack.herokuapp.com/)
[![Chat](https://stakater.github.io/README/stakater-chat-btn.png)](https://stakater.slack.com/messages/CA66MMYSE/)

## Extending

If you'd like to extend the functionality of Ingress Monitor Controller, please refer to the documentation
 [here](docs/developing/extension.md)

## Testing

For running tests, please refer to the documentation [here](docs/developing/testing.md)

## Contributing

If you'd like to contribute any fixes or enhancements, please refer to the documentation
 [here](docs/developing/contributing.md)

## Bug Reports & Feature Requests

Please use the [issue tracker](https://github.com/stakater/IngressMonitorController/issues) to report any bugs or file feature requests.

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
