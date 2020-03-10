ingressmonitorcontroller
========================
IngressMonitorController chart that runs on kubernetes

Current chart version is `v1.0.92`

Source code can be found [here](https://github.com/stakater/IngressMonitorController)



## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| ingressMonitorController.config.labels.version | string | `"v1.0.92"` |  |
| ingressMonitorController.configFilePath | string | `"/etc/IngressMonitorController/config.yaml"` |  |
| ingressMonitorController.data."config.yaml" | string | `"providers:\n- name: UptimeRobot\n  apiKey: your-api-key\n  apiURL: https://google.com\n  alertContacts: some-alert-contacts\nenableMonitorDeletion: true\nmonitorNameTemplate: \"{{.Namespace}}-{{.IngressName}}\"\n# how often (in seconds) monitors should be synced to their Kubernetes resources (0 = disabled)\nresyncPeriod: 0\n# creationDelay is a duration string to add a delay before creating new monitor (e.g., to allow DNS to catch up first)\n# https://golang.org/pkg/time/#ParseDuration\ncreationDelay: 0"` |  |
| ingressMonitorController.deployment.annotations."configmap.reloader.stakater.com/reload" | string | `"ingressmonitorcontroller"` |  |
| ingressMonitorController.deployment.labels.version | string | `"v1.0.92"` |  |
| ingressMonitorController.existingSecret | string | `""` |  |
| ingressMonitorController.image.name | string | `"stakater/ingressmonitorcontroller"` |  |
| ingressMonitorController.image.pullPolicy | string | `"IfNotPresent"` |  |
| ingressMonitorController.image.tag | string | `"v1.0.92"` |  |
| ingressMonitorController.logFormat | string | `"text"` |  |
| ingressMonitorController.logLevel | string | `"info"` |  |
| ingressMonitorController.matchLabels.group | string | `"com.stakater.platform"` |  |
| ingressMonitorController.matchLabels.provider | string | `"stakater"` |  |
| ingressMonitorController.mount | string | `"configMap"` |  |
| ingressMonitorController.rbac.create | bool | `true` |  |
| ingressMonitorController.serviceAccount.create | bool | `true` |  |
| ingressMonitorController.serviceAccount.labels.version | string | `"v1.0.92"` |  |
| ingressMonitorController.serviceAccount.name | string | `"ingressmonitorcontroller"` |  |
| ingressMonitorController.tolerations | object | `{}` |  |
| ingressMonitorController.useFullName | bool | `false` |  |
| ingressMonitorController.watchNamespace | string | `""` |  |
| kubernetes.host | string | `"https://kubernetes.default"` |  |
