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

### OpenStack

Due to [a bug](https://github.com/operator-framework/operator-sdk/issues/4684),
installation on OpenStack requires using an image of [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy)
provided by Red Hat:

```
helm install --set kube-rbac-proxy.image.repository=registry.redhat.io/openshift4/ose-kube-rbac-proxy,kube-rbac-proxy.image.tag=v4.7.0 stakater/ingressmonitorcontroller
```

## Chart Values

| Key                              | Default                                   | Description                                                                                    |
| -------------------------------- | ----------------------------------------- | ---------------------------------------------------------------------------------------------- |
| global.labels                    | ``                                        | Labels to be added to all components                                                           |
| replicaCount                     | `1`                                       | Replicas for operator                                                                          |
| image.repository                 | `"stakater/ingressmonitorcontroller"`     | Image repository                                                                               |
| image.tag                        | `LATEST_CHART_VERSION`                    | Tag of the Image                                                                               |
| image.pullPolicy                 | `IfNotPresent`                            | Pull policy for the image                                                                      |
| imagePullSecrets                 | ``                                        | List of secrets used to pull images                                                            |
| nameOverride                     | `""`                                      | Partial override for ingress-monitor-controller.fullname template (will keep the release name) |
| fullnameOverride                 | `""`                                      | Full override for ingress-monitor-controller.fullname template                                 |
| watchNamespaces                  | `""`                                      | Comma separated namespace names, set empty to watch all namespaces                             |
| configSecretName                 | `"imc-config"`                            | Name of secret that contains configuration                                                     |
| rbac.create                      | `true`                                    | Whether to create RBAC                                                                         |
| rbac.allowProxyRole              | `true`                                    | Whether to create RBAC for proxy                                                               |
| rbac.allowMetricsReaderRole      | `true`                                    | Whether to create RBAC for metrics-reader                                                      |
| rbac.allowLeaderElectionRole     | `true`                                    | Whether to create leader-election                                                              |
| serviceAccount.create            | `true`                                    | Whether to create serviceAccount                                                               |
| serviceAccount.name              | `""`                                      | Name for ServiceAccount, if empty the default chart name will be used                          |
| serviceAccount.labels            | `{}`                                      | Additional labels on ServiceAccount                                                            |
| serviceAccount.annotations       | `{}`                                      | Additional annotations on ServiceAccount                                                       |
| serviceMonitor.enabled           | `false`                                   | Create ServiceMonitor for metrics                                                              |
| podAnnotations                   | `""`                                      | Additional annotations on deployment                                                           |
| resources                        | `{}`                                      | Requests/Limits for operator                                                                   |
| securityContext                  | `{}`                                      | Override for SecurityContext                                                                   |
| podSecurityContext               | `{}`                                      | Override for deployment.Spec.securityContext                                                   |
| nodeSelector                     | `{}`                                      | Override for NodeSelector                                                                      |
| tolerations                      | `{}`                                      | Override for Tolerations                                                                       |
| affinity                         | `{}`                                      | Override for Affinity                                                                          |
| kube-rbac-proxy.image.repository | `"http://quay.io/brancz/kube-rbac-proxy"` | Image repository                                                                               |
| kube-rbac-proxy.image.tag        | `v0.1.0`                                  | Tag of the Image                                                                               |
| kube-rbac-proxy.image.pullPolicy | `IfNotPresent`                            | Pull policy for the image                                                                      |
