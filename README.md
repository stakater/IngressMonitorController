# ![](assets/web/IMC-round-100px.png) Ingress Monitor Controller

A Kubernetes/Openshift controller to watch ingresses/routes and create liveness alerts for your apps/microservices in Uptime checkers.

[![Get started with Stakater](https://stakater.github.io/README/stakater-github-banner.png)](http://stakater.com/?utm_source=IngressMonitorController&utm_medium=github)

## Problem Statement

We want to monitor ingresses in a kubernetes cluster and routes in openshift cluster via any uptime checker but the problem is having to manually check for new ingresses or routes / removed ingresses or routes and add them to the checker or remove them.

## Solution

This controller will continuously watch ingresses/routes in specific or all namespaces, and automatically add / remove monitors
 in any of the uptime checkers. With the help of this solution, you can keep a check on your services and see whether
  they're up and running and live, without worrying about manually registering them on the Uptime checker.

## Supported Uptime Checkers

Currently we support the following monitors:

- [UptimeRobot](https://uptimerobot.com) ([Additional Config](docs/uptimerobot-configuration.md))
- [Pingdom](https://pingdom.com) ([Additional Config](docs/pingdom-configuration.md)) (Not fully tested)
- [StatusCake](https://www.statuscake.com) ([Additional Config](docs/statuscake-configuration.md))
- [Uptime](http://uptime.com) ([Additional Config](docs/uptime-configurations.md))
- [Updown](https://updown.io/)([Additional Config](docs/updown-configuration.md))

## Usage

The following quickstart let's you set up Ingress Monitor Controller to register uptime monitors for ingresses/routes in all namespaces:

1. Download the
 [manifest file](https://raw.githubusercontent.com/stakater/IngressMonitorController/master/deployments/kubernetes/ingressmonitorcontroller.yaml)

2. Open the downloaded file in a text editor. Configure the uptime checker in the `config.yaml` data for the ConfigMap resource, and set the following properties

   | Key           | Description                                                                                                                      |
   | ------------- | -------------------------------------------------------------------------------------------------------------------------------- |
   | name          | Name of the provider (e.g. UptimeRobot)                                                                                          |
   | apiKey        | ApiKey of the provider                                                                                                           |
   | apiURL        | Base url of the ApiProvider with trailing slash (e.g. https://api.uptimerobot.com/v2/)                                           |
   | alertContacts | A `-` separated list of contact id's inside double quotes that you want to add to the monitors (e.g. "1234567_8_9-9876543_2_1" ) |
   | creationDelay | A duration string (e.g., "300ms", "1h50s") to delay the creation of a monitor. (default: 0)                                      |

    *Note:* Follow [this](docs/uptimerobot-configuration.md) guide to see how to fetch `alertContacts` from UptimeRobot.

    *Note 1:* See the section `Using Secrets` [here](docs/Deploying-to-Kubernetes.md) if you do not want to use                  ConfigMap (because API-Key in plain text) and want to use Secrets to hide API keys 
             

3. Enable for your Ingress/Route

   You will need to add the following annotation on your ingresses/routes so that the controller is able to recognize and monitor it.

   ```yaml
   "monitor.stakater.com/enabled": "true"
   ```
4. Deploy the controller by running the following command:

    For Kubernetes Cluster
   ```bash
   kubectl apply -f ingressmonitorcontroller.yaml -n default
   ```
   For Openshift Cluster
   ```bash
   oc create -f ingressmonitorcontroller.yaml -n default
   ```

### Helm Charts

Alternatively if you have configured helm on your cluster, you can add ingressmonitorcontroller to helm from helm's public chart repository and deploy it via helm using below mentioned commands

 ```bash
helm repo add stable https://kubernetes-charts.storage.googleapis.com/

helm repo update

helm install stable/ingressmonitorcontroller
```

## Help

### Documentation
You can find more detailed documentation for configuration, extension, and support for other Uptime checkers etc. [here](docs/Deploying-to-Kubernetes.md)

### Contributig

If you'd like to contribute any fixes or enhancements, please refer to the documentation [here](CONTRIBUTING.md)

### Have a question?
File a GitHub [issue](https://github.com/stakater/IngressMonitorController/issues), or send us an [email](mailto:hello@stakater.com).

### Talk to us on Slack
Join and talk to us on the #tools-ingressmonitor channel for discussing the Ingress Monitor Controller

[![Join Slack](https://stakater.github.io/README/stakater-join-slack-btn.png)](https://stakater-slack.herokuapp.com/)
[![Chat](https://stakater.github.io/README/stakater-chat-btn.png)](https://stakater.slack.com/messages/CA66MMYSE/)

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

