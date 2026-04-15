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

helm install stakater/ingressmonitorcontroller
```

## Chart Values

| Key                                              | Default                               | Description                                                                                       |
| ------------------------------------------------ | ------------------------------------- | ------------------------------------------------------------------------------------------------- |
| global.labels                                    | ``                                    | Labels to be added to all components                                                              |
| replicaCount                                     | `1`                                   | Replicas for operator                                                                             |
| image.name                                       | `"stakater/ingressmonitorcontroller"` | Image repository                                                                                  |
| image.tag                                        | `LATEST_CHART_VERSION`                | Tag of the Image                                                                                  |
| image.pullPolicy                                 | `Always`                              | Pull policy for the image                                                                         |
| imagePullSecrets                                 | ``                                    | List of secrets used to pull images                                                               |
| nameOverride                                     | `""`                                  | Partial override for ingress-monitor-controller.fullname template (will keep the release name)    |
| fullnameOverride                                 | `""`                                  | Full override for ingress-monitor-controller.fullname template                                    |
| watchNamespaces                                  | `""`                                  | Comma separated namespace names, set empty to watch all namespaces                                |
| configSecretName                                 | `"imc-config"`                        | Name of secret that contains configuration                                                        |
| rbac.create                                      | `true`                                | Whether to create RBAC                                                                            |
| rbac.allowProxyRole                              | `true`                                | Whether to create RBAC for proxy                                                                  |
| rbac.allowMetricsReaderRole                      | `true`                                | Whether to create RBAC for metrics-reader                                                         |
| rbac.allowLeaderElectionRole                     | `true`                                | Whether to create leader-election                                                                 |
| serviceAccount.create                            | `true`                                | Whether to create serviceAccount                                                                  |
| serviceAccount.name                              | `""`                                  | Name for ServiceAccount, if empty the default chart name will be used                             |
| serviceAccount.labels                            | `{}`                                  | Additional labels on ServiceAccount                                                               |
| serviceAccount.annotations                       | `{}`                                  | Additional annotations on ServiceAccount                                                          |
| serviceMonitor.enabled                           | `false`                               | Create ServiceMonitor for metrics                                                                 |
| serviceMonitor.prometheusServiceAccountName      | `""`                                  | Name of the Prometheus ServiceAccount to bind to the metrics-reader ClusterRole                   |
| serviceMonitor.prometheusServiceAccountNamespace | `""`                                  | Namespace of the Prometheus ServiceAccount                                                        |
| certManager.enabled                              | `false`                               | Create cert-manager resources (Issuer/Certificate) for securing the metrics endpoint              |
| certManager.selfSigned                           | `true`                                | Create a self-signed Issuer; set to `false` to use an existing issuer via `certManager.issuerRef` |
| certManager.injectCA                             | `true`                                | Inject the issuing CA bundle into ServiceMonitor; set to `false` for ACME/public issuers          |
| certManager.issuerRef.name                       | `""`                                  | Name of an existing Issuer or ClusterIssuer (used when `selfSigned` is `false`)                   |
| certManager.issuerRef.kind                       | `ClusterIssuer`                       | Kind of the referenced issuer resource (`Issuer` or `ClusterIssuer`)                              |
| certManager.issuerRef.group                      | `cert-manager.io`                     | API group of the referenced issuer resource                                                       |
| certManager.duration                             | `"2160h"`                             | Requested certificate duration (e.g. `2160h` = 90 days)                                           |
| certManager.renewBefore                          | `"360h"`                              | How long before expiry cert-manager should renew the certificate                                  |
| podAnnotations                                   | `""`                                  | Additional annotations on deployment                                                              |
| podLabels                                        | `{}`                                  | Additional labels for the Pod template                                                            |
| resources                                        | `{}`                                  | Requests/Limits for operator                                                                      |
| securityContext                                  | `{}`                                  | Override for SecurityContext                                                                      |
| podSecurityContext                               | `{}`                                  | Override for deployment.Spec.securityContext                                                      |
| nodeSelector                                     | `{}`                                  | Override for NodeSelector                                                                         |
| tolerations                                      | `{}`                                  | Override for Tolerations                                                                          |
| affinity                                         | `{}`                                  | Override for Affinity                                                                             |
