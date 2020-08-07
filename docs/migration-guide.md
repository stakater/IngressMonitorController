# Migrating to Operator

## What has changed

IMC controller worked on the basis of annotations on the routes or ingresses. That is obsolete now and the whole lifecycle
and additional configuration for monitors are now managed by `EndpointMonitor` custom resource. Provider specific annotations
are now part of Custom Resource now:

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: uptimerobot-config-example
spec:
  forceHttps: true
  providers: "UptimeRobot"
  healthEndpoint: "/healthzzz"
  urlFrom:
    routeRef:
      name: frontend
  uptimeRobotConfig:
    interval: 600
    monitorType: "http"
    KeywordExists: "yes"
    KeywordValue: "404"
```

Old way of using it with Controller:

```yaml
apiVersion: v1
kind: Route
metadata:
  name: frontend
  annotations:
    monitor.stakater.com/enabled: true
    uptimerobot.monitor.stakater.com/interval: 600
    uptimerobot.monitor.stakater.com/monitor-type: 'http'
    uptimerobot.monitor.stakater.com/keyword-exists: 'yes'
    uptimerobot.monitor.stakater.com/keyword-value: '404'
spec:
  to:
    kind: Service
    name: hello-openshift
  port:
    targetPort: http
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
  wildcardPolicy: None
```

## Migration Guideline

**WIP** Create CR for all annotated routes/ingresses
