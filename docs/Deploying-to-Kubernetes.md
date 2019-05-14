# Deploying to Kubernetes

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
| secrets               | An array of secrets that you want to mount to your controller                 |
| enableMonitorDeletion | A safeguard flag that is used to enable or disable monitor deletion on ingress deletion (Useful for prod environments where you don't want to remove monitor on ingress deletion) |
| resyncPeriod          | Resync period in seconds, allows to re-sync periodically the monitors with the Ingresses. Defaults to 0 (= disabled) |
| watchNamespace        | Name of the namespace if you want to monitor ingresses only in that namespace. Defaults to null |
| mount                 | `"secret"` or `"configMap"`. How to pass user credentials, API keys etc.          |  

For the list of providers, there are a number of options that you need to specify. The table below lists them:

| Key           | Description                                                               |
|---------------|---------------------------------------------------------------------------|
| name          | Name of the provider (From the list of supported uptime checkers)         |
| apiKey        | ApiKey of the provider                                                    |
| apiURL        | Base url of the ApiProvider                                               |
| alertContacts | A `-` separated list of contact id's that you want to add to the monitors |

#### Uptime Checker specific
##### UptimeRobot ([https://uptimerobot.com](https://uptimerobot.com))
Follow the [UptimeRobot Configuration guide](uptimerobot-configuration.md) to see how to fetch `alertContacts` from UptimeRobot.

##### Pingdom ([https://pingdom.com](https://pingdom.com))
[Pingdom Configuration guide](../docs/pingdom-configuration.md)

##### Statuscake ([https://www.statuscake.com](https://www.statuscake.com))
[Statuscake Configuration guide](../docs/statuscake-configuration.md)

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

##### Using Secrets

To pass user credentials/ API keys in secrets:
    
  1. Open [values.yaml](https://github.com/stakater/IngressMonitorController/blob/master/deployments/kubernetes/chart/ingressmonitorcontroller/values.yaml) file by navigating to `deployments/kubernetes/chart/ingressmonitorcontroller/`
  
  2. Set `mount` equals to `"secret"` and pass the data in the data section at the bottom.
  
  3. Run `helm template . > file_to_generate.yaml`
  
  4. Deploy using the `Deploying` section below.

##### Using ConfigMap

To pass user credentials/ API keys in secrets:
     
  1. Open [values.yaml](https://github.com/stakater/IngressMonitorController/blob/master/deployments/kubernetes/chart/ingressmonitorcontroller/values.yaml) file by navigating to `deployments/kubernetes/chart/ingressmonitorcontroller/`
  
  2. Set `mount` equals to `"configMap"` and pass the data in the data section at the bottom.
  
  3. Run `helm template . > file_to_generate.yaml`
  
  4. Deploy using the `Deploying` section below.

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

   i. `helm fetch --untar stakater/ingressmonitorcontroller`

   ii. Open and edit `ingressmonitorcontroller/values.yaml` in a text editor
3. Install the chart
   * `helm install stakater/ingressmonitorcontroller -f ingressmonitorcontroller/values.yaml -n ingressmonitorcontroller`

##### Running for a single namespace

Set `watchNamespace` to `<namespace-name>` in `ingressmonitorcontroller/values.yaml` before applying the helm chart
 and the controller will then watch in that namespace.

## Bug Reports & Feature Requests

Please use the [issue tracker](https://github.com/stakater/IngressMonitorController/issues) to report any bugs or file feature requests.

## Changelog

View our closed [Pull Requests](https://github.com/stakater/IngressMonitorController/pulls?q=is%3Apr+is%3Aclosed).
