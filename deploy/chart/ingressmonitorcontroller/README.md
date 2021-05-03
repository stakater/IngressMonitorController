# IngressMonitorController Helm Chart


Helm chart for the new IngressMonitorController operator that runs on kubernetes

This chart supports v2.x.x as IMC has been shifted to operator from that version. For controller based approach, refer to [release-v1](https://github.com/stakater/IngressMonitorController/tree/release-v1/deployments/kubernetes)

Source code can be found [here](https://github.com/stakater/IngressMonitorController)

## Installation

To install IMC helm chart run the following

```sh
# Install CRDs
kubectl apply -f https://raw.githubusercontent.com/stakater/IngressMonitorController/master/deploy/crds/endpointmonitor.stakater.com_endpointmonitors_crd.yaml

# Install Chart
helm repo add stakater https://stakater.github.io/stakater-charts

helm repo update

helm install stakater/ingressmonitorcontroller
```

## Chart Values

| Key | Default | Description |
|-----|---------|-------------|
| global.labels | `` | Labels to be added to all components |
| watchNamespace | `` | Whether to watch any single namespace, set empty to watch all namespaces |
| useFullName | `false` |  |
| deployment.annotations | `"configmap.reloader.stakater.com/reload": "ingressmonitorcontroller"` |  Additional annotations on deployment |
| deployment.labels | `` | Additional labels on deployment |
| deployment.replicas | `1` | Replicas for deployment |
| deployment.revisionHistoryLimit | `2` | Limit on rollout history  |
| deployment.operatorName | `ingressmonitorcontroller` |  |
| deployment.logLevel | `info` | Log Level |
| deployment.logFormat | `text` | Formatting logs |
| deployment.image.name | `"stakater/ingressmonitorcontroller"` | Image repository |
| deployment.image.tag | `LATEST_CHART_VERSION` | Tag of the Image |
| deployment.image.pullPolicy | `Always` | Pull policy for the image |
| rbac.create | `true` | Whether to create RBAC (Role/Cluster & RoleBinding/ClusterRoleBinding) |
| rbac.serviceAccount.create | `true` | Whether to create serviceAccount |
| rbac.serviceAccount.name | `""` | Name for ServiceAccount, if empty the default chart name will be used |
| rbac.serviceAccount.labels | `{}` | Additional labels on ServiceAccount |
| rbac.serviceAccount.annotations | `{}` | Additional annotations on ServiceAccount|
| secret.useExisting | `false` | Whether to use an existing secret, if true, this chart will not create secret |
| secret.name | `""` | Name used for secret, either already existing secret or created from this chart, if empty the default chart name will be used |
| secret.labels | `{}` | Additional labels on Secret |
| secret.annotations | `{}` | Additional annotations on Secret|
| secret.data."config.yaml" | `"providers:\n- name: UptimeRobot\n  apiKey: your-api-key\n  apiURL: https://google.com\n  alertContacts: some-alert-contacts\nenableMonitorDeletion: true\nmonitorNameTemplate: \"{{.Namespace}}-{{.IngressName}}\"\n# how often (in seconds) monitors should be synced to their Kubernetes resources (0 = disabled)\nresyncPeriod: 0\n# creationDelay is a duration string to add a delay before creating new monitor (e.g., to allow DNS to catch up first)\n# https://golang.org/pkg/time/#ParseDuration\ncreationDelay: 0"` | Config for secret used for IMC |

