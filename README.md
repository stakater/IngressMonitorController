# Ingress Monitor Controller

## Problem Statement

We want to monitor ingresses in a kubernetes cluster via any uptime checker but the problem is to manually check for new ingresses / removed ingresses and add them to the checker or remove them. There isn't any out of the box solution for this as of now.

## Solution

This controller will continuously watch ingresses in the namespace it is running, and automatically add / remove monitors in any of the uptime checkers. With the help of this solution, you can keep a check on your services and see whether they're up and running and live.

## Supported Uptime Checkers

Currently we support the following monitors:

- [UptimeRobot](https://uptimerobot.com)

## Deploying to Kubernetes

You have to first clone or download the repository contents. The kubernetes deployment and files are provided inside `kubernetes-manifests` folder.

### Enabling

By default, the controller ignores the ingresses without a specific annotation on it. You will need to add the following annotation on your ingresses so that the controller is able to recognize and monitor the ingresses.

```yaml
"monitor.stakater.com/enabled": "true"
```

The annotation key is `monitor.stakater.com/enabled` and you can toggle the value of this annotation between `true` and `false` to enable or disable monitoring of that specific ingress.

### Configuring

First of all you need to modify `configmap.yaml`'s `config.yaml` file. Following are the available options that you can use to customize the controller:

| Key                   |Description                                                                    |
|-----------------------|-------------------------------------------------------------------------------|
| providers             | An array of uptime providers that you want to add to your controller          |
| enableMonitorDeletion | A flag that is used to enable or disable monitor deletion on ingress deletion |

For the list of providers, there's a number of options that you need to specify. The table below lists them:

| Key           | Description                                                               |
|---------------|---------------------------------------------------------------------------|
| name          | Name of the provider (From the list of supported uptime checkers)         |
| apiKey        | ApiKey of the provider                                                    |
| apiURL        | Base url of the ApiProvider                                               |
| alertContacts | A `-` separated list of contact id's that you want to add to the monitors |

### Deploying

You can deploy the controller in the namespace you want to monitor by running the following kubectl commands:

```bash
kubectl apply -f configmap.yaml -n <namespace>
kubectl apply -f rbac.yaml -n <namespace>
kubectl apply -f deployment.yaml -n <namespace>
```

## Adding support for a new Monitor

You can easily implement a new monitor and use it via the controller. First of all, you will need to create a new service struct that implements the following monitor service interface

```go
type MonitorService interface {
	GetAll() []Monitor
	Add(m Monitor)
	Update(m Monitor)
	GetByName(name string) (*Monitor, error)
	Remove(m Monitor)
	Setup(apiKey string, url string, alertContacts string)
}
```

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