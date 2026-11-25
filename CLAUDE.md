# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

IngressMonitorController is a Kubernetes/OpenShift operator (built with Kubebuilder v4 / operator-sdk) that reconciles an `EndpointMonitor` custom resource (group `endpointmonitor.stakater.com`, version `v1alpha1`) and creates/updates/deletes uptime monitors against external SaaS uptime checkers (UptimeRobot, Pingdom, Pingdom Transaction, StatusCake, Uptime.com, Updown, Azure Application Insights, Google Cloud Monitoring, Grafana Synthetic Monitoring).

The Go module is `github.com/stakater/IngressMonitorController/v2`. Note the `v2` suffix — any new import paths must include it.

## Common commands

All built through the Makefile. Tool binaries (controller-gen, kustomize, setup-envtest, golangci-lint) are auto-installed into `./bin/` on first use; do not install them globally.

| Task | Command |
| --- | --- |
| Build manager binary (`bin/manager`) | `make build` |
| Run controller against the current kubecontext | `OPERATOR_NAMESPACE=test make run` |
| Generate DeepCopy code (from kubebuilder markers) | `make generate` |
| Regenerate CRDs / RBAC / webhook manifests | `make manifests` |
| Regenerate CRDs into the Helm chart | `make generate-crds` |
| `go fmt` / `go vet` | `make fmt` / `make vet` |
| Lint (golangci-lint v1.64) | `make verify-golangci-lint` |
| Run full test suite under envtest (K8s 1.31) | `make test` |
| Run a single package/test | `KUBEBUILDER_ASSETS="$(./bin/setup-envtest use 1.31.0 --bin-dir ./bin -p path)" go test ./pkg/monitors/uptimerobot/ -run TestName -v` |
| Apply CRDs to current cluster | `make install` |
| Deploy operator via kustomize | `make deploy IMG=<image>` |
| Build container image | `make docker-build IMG=<image>` |
| Render consolidated install YAML to `dist/install.yaml` | `make build-installer` |
| OLM bundle | `make bundle` / `make bundle-build` |
| Bump Helm chart version | `make bump-chart VERSION=x.y.z` |

`make test` and `make build` both invoke `manifests generate fmt vet` first, so a code change that touches kubebuilder markers or `*_types.go` will trigger regeneration automatically. CI uses `golangci-lint v1.64` with `--timeout 10m`; match this locally.

### Running tests against real providers

Provider monitor tests (`pkg/monitors/<provider>/*_test.go`) talk to real SaaS APIs. They read configuration from a YAML file pointed at by `CONFIG_FILE_PATH` (defaults to `../../../.local/test-config.yaml` — i.e. a `.local/test-config.yaml` at repo root). Sample shapes live under `examples/configs/`. Tests also expect a `test` namespace to exist in the active kube cluster. See `CONTRIBUTING.md` for the full setup recipe; constants like `correctTestAPIKey` may need to be edited to match your tenant.

## High-level architecture

### Entry point

`cmd/main.go` builds a controller-runtime Manager:
- Detects OpenShift at init time via `pkg/kube.IsOpenshift` (probes `/apis/route.openshift.io`) and conditionally registers the `routev1` scheme. This means **the binary needs cluster reachability at startup** to detect environment — keep this in mind when running locally.
- Reads `WATCH_NAMESPACE` (comma-separated). Empty → cluster-scoped cache; non-empty → a `MultiNamespacedCache`. The empty-vs-`{"":{}}` distinction is load-bearing (see comment at `cmd/main.go:144`) — do not "normalize" it.
- Loads the controller config from a Kubernetes Secret (`imc-config` by default, overridable via `CONFIG_SECRET_NAME`) in the operator's namespace, via `config.LoadControllerConfig` → `secret.LoadSecretData`. The secret must contain a `config.yaml` key.
- Calls `monitors.SetupMonitorServicesForProviders(cfg.Providers)` to instantiate one `MonitorServiceProxy` per configured provider, then wires them into `EndpointMonitorReconciler`.

### Reconcile loop

`internal/controller/endpointmonitor_controller.go` is the single controller (`For(&EndpointMonitor{})`). Reconcile flow:
1. Compute the external monitor name from `MonitorNameTemplate` (Go `html/template`, defaults to `{{.Name}}-{{.Namespace}}`). Template parsing is done via `pkg/util.GetNameTemplateFormat`, which converts the template into a `fmt.Sprintf` format string with two ordered args.
2. `GetMonitorOfType(spec)` picks the right `MonitorServiceProxy` by inspecting which `*Config` field is set on the spec (first non-nil wins; otherwise falls back to the first configured provider). The dispatch order matters when multiple configs are set on a single CR — Pingdom Transaction is checked before Pingdom, etc.
3. If the CR is gone → `handleDelete` (gated by `EnableMonitorDeletion`). Delete iterates **all** monitor services because the monitor may exist in providers other than the one currently selected by spec.
4. If the monitor exists by name in the provider → `handleUpdate` (only calls `Update` when `monitorService.Equal(old, new)` is false).
5. Otherwise → `handleCreate`, but respect `CreationDelay` by returning `RequeueAfter: delay` when the CR is younger than the delay.
6. Returns `RequeueAfter: ReconciliationRequeueTime` (from env `REQUEUE_TIME`, default 300s).

The three handler files (`endpointmonitor_created.go`, `endpointmonitor_updated.go`, `endpointmonitor_deleted.go`) are intentionally tiny — extend them rather than putting logic in `Reconcile`.

### URL resolution

`pkg/kube/util/url.go::GetMonitorURL` decides what URL the monitor points at:
- `spec.url` literal wins if set.
- Otherwise resolves via `spec.urlFrom.ingressRef` (works on both K8s + OpenShift) or `spec.urlFrom.routeRef` (OpenShift only — silently errors on vanilla K8s).
- StatusCake `Heartbeat` test type is special-cased to an empty URL.
- Ingress/Route URL extraction lives in `pkg/kube/wrappers/`, applies `forceHttps` and `healthEndpoint`.

### Provider plugin architecture

`pkg/monitors/` is the extension point. The contract:

```go
// pkg/monitors/monitor-service.go
type MonitorService interface {
    GetAll() ([]models.Monitor, error)
    Add(m models.Monitor)
    Update(m models.Monitor)
    GetByName(name string) (*models.Monitor, error)
    Remove(m models.Monitor)
    Setup(p config.Provider)
    Equal(oldMonitor, newMonitor models.Monitor) bool
}
```

`MonitorServiceProxy` (`pkg/monitors/monitor-proxy.go`) is the polymorphic wrapper. **To add a new provider you must touch all of these:**
1. Create `pkg/monitors/<name>/` with a struct implementing `MonitorService`. Convention: a `<name>-monitor.go`, a `<name>-mappers.go` for API↔`models.Monitor` translation, and `<name>-responses.go` for API DTOs.
2. Add a `Type<Name>` constant + `case` in `MonitorServiceProxy.OfType` (constructs the impl) and in `MonitorServiceProxy.ExtractConfig` (extracts the spec sub-struct).
3. Add a new `*<Name>Config` field in `EndpointMonitorSpec` (`api/v1alpha1/endpointmonitor_types.go`), then add a branch in `EndpointMonitorReconciler.GetMonitorOfType` so the controller can pick it.
4. Add provider-level config fields to `config.Provider` (`pkg/config/config.go`) if the SaaS needs apiKey/auth/etc.
5. Run `make manifests generate generate-crds` and commit the regenerated CRDs (under `config/crd/bases/` and `charts/ingressmonitorcontroller/crds/`) and `zz_generated.deepcopy.go`.

`SetupMonitorServicesForProvidersTest` (same file) intentionally restricts the providers wired up in tests to UptimeRobot + StatusCake — extend the `allowedProviders` slice if a new provider has integration tests.

### Configuration model

Two layers of config:
- **Operator-level** (`config.Config` from the Secret): provider credentials, `enableMonitorDeletion`, `monitorNameTemplate`, `creationDelay`, `resyncPeriod`. Loaded once at startup into the package-level `IngressMonitorControllerConfig`; treat as immutable post-boot.
- **Per-CR** (`EndpointMonitorSpec.<provider>Config`): per-monitor knobs (interval, alertContacts overrides, monitor type, etc.). Passed through `MonitorServiceProxy.ExtractConfig` to the provider implementation untyped (`interface{}`) — providers type-assert to their own config struct.

### Notable environment variables

| Var | Default | Effect |
| --- | --- | --- |
| `WATCH_NAMESPACE` | required when set; empty/unset → cluster scope | Comma-separated namespace list |
| `OPERATOR_NAMESPACE` | autodetected via downward API | Used to locate the config Secret; **must be set when running `make run` locally** |
| `CONFIG_SECRET_NAME` | `imc-config` | Name of the Secret holding `config.yaml` |
| `REQUEUE_TIME` | `300` (seconds) | Reconcile requeue interval |
| `CONFIG_FILE_PATH` | `../../../.local/test-config.yaml` | Test-time substitute for the Secret-backed config |

## Repository layout notes

- `api/v1alpha1/` — CRD types. **Generated code (`zz_generated.deepcopy.go`) lives here too; do not hand-edit.**
- `internal/controller/` — controller + reconcile handlers.
- `cmd/main.go` — manager bootstrap.
- `pkg/monitors/<provider>/` — one directory per provider; this is where provider-specific logic belongs.
- `pkg/monitors/monitor-proxy.go` + `monitor-service.go` — dispatch layer; edit when adding providers.
- `pkg/kube/` — environment detection, ingress/route URL wrappers.
- `pkg/config/` — operator config types + Secret loader.
- `pkg/secret/`, `pkg/http/`, `pkg/util/` — small shared helpers.
- `config/` — Kustomize manifests (CRDs, RBAC, manager, default overlay). Generated by `make manifests`; edits to `config/crd/bases/` get clobbered — change the kubebuilder markers in `*_types.go` instead.
- `charts/ingressmonitorcontroller/` — Helm chart shipped to the `stakater-charts` repo. CRDs under `charts/.../crds/` are generated by `make generate-crds`. Version bumped via `make bump-chart VERSION=x.y.z`.
- `bundle/`, `bundle.Dockerfile` — OLM bundle artifacts.
- `examples/` — sample `EndpointMonitor` CRs and provider config YAMLs (also used as test config templates).
- `docs/` — per-provider configuration guides; user-facing.

## Conventions that aren't obvious from the code

- **Release cuts** happen via tag push: `git tag v<VERSION> && git push origin v<VERSION>` (see README). GoReleaser config in `.goreleaser.yml`.
- Recent commits use `[skip-ci] Update artifacts` for auto-generated CRD/chart updates — don't be surprised by them in `git log`.
- The default branch is `master`, not `main`.
- Go version pinned to `1.24` (`go.mod`, Dockerfile). `make build` targets distroless static, CGO disabled.
- When a `spec` has multiple `*Config` blocks set, only one wins per reconcile — see `GetMonitorOfType` for the priority order. If a user expects fan-out to multiple providers, that requires multiple `EndpointMonitor` CRs (one per provider).
