# ![](assets/web/IMC-round-100px.png) Ingress Monitor Controller

### DEPRECATION NOTICE: 

**IMC has now been converted to an Operator and we have stopped support from our side for the controller based implementation
, although support from community for the controller is still appreciated. Using Operator is recommended and existing users can follow
[Migration To Operator](./docs/migration-guide.md) for migrating to Operator. Although, Controller based implementation is maintained at [release-v1](https://github.com/stakater/IngressMonitorController/tree/release-v1) instead.**


An operator to watch ingresses/routes and create liveness alerts for your apps/microservices in Uptime checkers.

[![Get started with Stakater](https://stakater.github.io/README/stakater-github-banner.png)](http://stakater.com/?utm_source=IngressMonitorController&utm_medium=github)

## Problem Statement

We want to monitor ingresses in a kubernetes cluster and routes in openshift cluster via any uptime checker but the problem is having to manually check for new ingresses or routes / removed ingresses or routes and add them to the checker or remove them.

## Solution

This operator will continuously watch ingresses/routes based on defined `EndpointMonitor` custom resource, and 
automatically add / remove monitors in any of the uptime checkers. With the help of this solution, you can keep a check
 on your services and see whether they're up and running and live, without worrying about manually registering them on
  the Uptime checker.

## Supported Uptime Checkers

Currently we support the following monitors:

- [UptimeRobot](https://uptimerobot.com) ([Additional Config](docs/uptimerobot-configuration.md))
- [Pingdom](https://pingdom.com) ([Additional Config](docs/pingdom-configuration.md)) (Not fully tested)
- [StatusCake](https://www.statuscake.com) ([Additional Config](docs/statuscake-configuration.md))
- [Uptime](http://uptime.com) ([Additional Config](docs/uptime-configurations.md))
- [Updown](https://updown.io/) ([Additional Config](docs/updown-configuration.md))
- [Application Insights](https://docs.microsoft.com/en-us/azure/azure-monitor/app/monitor-web-app-availability) ([Additional Config](docs/appinsights-configuration.md))
- [gcloud](https://cloud.google.com/monitoring/uptime-checks) ([Additional Config](docs/gcloud-configuration.md))

## Usage

### Adding configuration

Configure the uptime checker configuration in the `config.yaml` based on your uptime provider. Add create a secret 
`imc-config` that holds `config.yaml` key:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: imc-config
data:
  config.yaml: >-
    <BASE64_ENCODED_CONFIG.YAML>
type: Opaque
```

#### Configuration Parameters

Following are the available options that you can use to customize the controller:

| Key                   |Description                                                                    |
|-----------------------|-------------------------------------------------------------------------------|
| providers             | An array of uptime providers that you want to add to your controller          |
| enableMonitorDeletion | A safeguard flag that is used to enable or disable monitor deletion on ingress deletion (Useful for prod environments where you don't want to remove monitor on ingress deletion) |
| resyncPeriod          | Resync period in seconds, allows to re-sync periodically the monitors with the Routes. Defaults to 0 (= disabled) |
| creationDelay        | CreationDelay is a duration string to add a delay before creating new monitor (e.g., to allow DNS to catch up first) |
| monitorNameTemplate    | Template for monitor name eg, `{{.Namespace}}-{{.Name}}`          |  


- Replace `BASE64_ENCODED_CONFIG.YAML` with your config.yaml file that is encoded in base64. 
- For detailed guide for the configuration refer to [Docs](./docs) and go through configuration guidelines for your uptime provider.
- For sample `config.yaml` files refer to [Sample Configs](examples/configs).
- Name of secret can be changed by setting environment variable `CONFIG_SECRET_NAME`.

### Add EndpointMonitor

`EndpointMonitor` resource can be used to manage monitors on static urls or route/ingress references.

- Specifying url:

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHtpps: true
  url: https://stakater.com
```

- Specifying route reference:

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: frontend
spec:
  forceHtpps: true
  urlFrom:
    routeRef:
      name: frontend
```

- Specifying ingress reference:

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: frontend
spec:
  forceHtpps: true
  urlFrom:
    ingressRef:
      name: frontend
```

NOTE: For provider specific additional configuration refer to [Docs](./docs) and go through configuration guidelines for your uptime provider.

## Deploying the Operator

The following quickstart let's you set up Ingress Monitor Controller to register uptime monitors for endpoints:

1) Clone this repository
```terminal
    $ git clone git@github.com:stakater/IngressMonitorController.git
```

2) Deploy dependencies(crds):
```terminal
    $ oc apply -f deploy/crds
```
 
3) Deploy ServiceAccount, Role, RoleBinding and Operator:
```terminal
   $ oc apply -f deploy/service_account.yaml
   $ oc apply -f deploy/role.yaml
   $ oc apply -f deploy/role_binding.yaml
   $ oc apply -f deploy/operator.yaml
```

### Environment Variables

| Key                   |Description                                                                    |
|-----------------------|-------------------------------------------------------------------------------|
| WATCH_NAMESPACE             | Use comma separated list of namespaces or leave the field empty to watch all namespaces(cluster scope)          |
| CONFIG_SECRET_NAME | Name of secret that holds the configuration |
| LOG_LEVEL          | Set logging level from debug,info,warn,error,fatal. Default value is Info |
| LOG_FORMAT        | Set logging format from text,json. Default value is text |


## Help

### Documentation

You can find more detailed documentation for configuration, extension, and support for other Uptime checkers etc. [here](docs/Deploying-to-Kubernetes.md)

### Contributing

If you'd like to contribute any fixes or enhancements, please refer to the documentation [here](CONTRIBUTING.md)

### Have a question?

File a GitHub [issue](https://github.com/stakater/IngressMonitorController/issues), or send us an [email](mailto:hello@stakater.com).

### Talk to us on Slack

Join and talk to us on the #tools-ingressmonitor channel for discussing the Ingress Monitor Controller

[![Join Slack](https://stakater.github.io/README/stakater-join-slack-btn.png)](https://slack.stakater.com/)
[![Chat](https://stakater.github.io/README/stakater-chat-btn.png)](https://stakater.slack.com/messages/CA66MMYSE/)

## License

Apache2 © [Stakater](http://stakater.com)

## About

The `IngressMonitorController` is maintained by [Stakater][website]. Like it? Please let us know at <hello@stakater.com>

See [our other projects][community]
or contact us in case of professional services and queries on <hello@stakater.com>

[website]: http://stakater.com/
[community]: https://www.stakater.com/projects-overview.html

The Google Cloud test infrastructure is sponsored by [JOSHMARTIN])(https://github.com/jshmrtn).

## Contributors

Stakater Team and the Open Source community! :trophy:
