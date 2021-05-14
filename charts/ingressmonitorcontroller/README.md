# IngressMonitorController Helm Chart

Helm chart for the new IngressMonitorController operator that runs on kubernetes

This chart supports v2.x.x as IMC has been shifted to operator from that version. For controller based approach, refer to [release-v1](https://github.com/stakater/IngressMonitorController/tree/release-v1/deployments/kubernetes)

Source code can be found [here](https://github.com/stakater/IngressMonitorController)

## Installation

To install IMC helm chart run the following

```sh

# Install Chart
helm repo add stakater https://stakater.github.io/stakater-charts

helm repo update

# Helm 2
helm install --set installCRDs=true stakater/ingressmonitorcontroller

# Helm 3
helm install stakater/ingressmonitorcontroller
```

## Chart Values

| Key                          | Default                               | Description                                                                                    |
| ---------------------------- | ------------------------------------- | ---------------------------------------------------------------------------------------------- |
| global.labels                | ``                                    | Labels to be added to all components                                                           |
| installCRDs                  | false                                 | Whether to install CRDs (Helm 2)                                                               |
| replicaCount                 | `1`                                   | Replicas for operator                                                                          |
| image.name                   | `"stakater/ingressmonitorcontroller"` | Image repository                                                                               |
| image.tag                    | `LATEST_CHART_VERSION`                | Tag of the Image                                                                               |
| image.pullPolicy             | `Always`                              | Pull policy for the image                                                                      |
| imagePullSecrets             | ``                                    | List of secrets used to pull images                                                            |
| nameOverride                 | `""`                                  | Partial override for ingress-monitor-controller.fullname template (will keep the release name) |
| fullnameOverride             | `""`                                  | Full override for ingress-monitor-controller.fullname template                                 |
| watchNamespaces              | `""`                                  | Comma separated namespace names, set empty to watch all namespaces                             |
| configSecretName             | `"imc-config"`                        | Name of secret that contains configuration                                                     |
| rbac.create                  | `true`                                | Whether to create RBAC                                                                         |
| rbac.allowProxyRole          | `true`                                | Whether to create RBAC for proxy                                                               |
| rbac.allowMetricsReaderRole  | `true`                                | Whether to create RBAC for metrics-reader                                                      |
| rbac.allowLeaderElectionRole | `true`                                | Whether to create leader-election                                                              |
| serviceAccount.create        | `true`                                | Whether to create serviceAccount                                                               |
| serviceAccount.name          | `""`                                  | Name for ServiceAccount, if empty the default chart name will be used                          |
| serviceAccount.labels        | `{}`                                  | Additional labels on ServiceAccount                                                            |
| serviceAccount.annotations   | `{}`                                  | Additional annotations on ServiceAccount                                                       |
| serviceMonitor.enabled       | `false`                               | Create ServiceMonitor for metrics                                                              |
| podAnnotations               | `""`                                  | Additional annotations on deployment                                                           |
| resources                    | `{}`                                  | Requests/Limits for operator                                                                   |
| securityContext              | `{}`                                  | Override for SecurityContext                                                                   |
| podSecurityContext           | `{}`                                  | Override for deployment.Spec.securityContext                                                   |
| nodeSelector                 | `{}`                                  | Override for NodeSelector                                                                      |
| tolerations                  | `{}`                                  | Override for Tolerations                                                                       |
| affinity                     | `{}`                                  | Override for Affinity                                                                          |
| env                          | `{}`                                  | Additional environment variables in the manager container                                      |
| envFrom                      | `{}`                                  | Additional sources to populate environment variables in the manager container                  |
