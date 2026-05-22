# IMC v3: Implementation Plan

Phased so each step compiles and is reviewable on its own. Tests added with each phase.

v3.0 ships **both** `ClusterMonitorProvider` (cluster-scoped) and `MonitorProvider` (namespaced) — multi-tenancy is a first-class v3.0 requirement, not a future addition.

**E2E test coverage scope:** the project has API credentials/test accounts for **UptimeRobot** and **StatusCake** only. Those two providers are validated end-to-end against a real upstream API on a kind cluster. The other seven providers (Pingdom, PingdomTransaction, Uptime, Updown, AppInsights, GCloud, Grafana) are **mock-tested only** — their `Setup` migration and reconciler wiring are covered by unit tests with mocked clients, but no live upstream call is exercised. This limitation must be stated explicitly in the v3 release notes and `docs/migration-v2-to-v3.md`.

## Phase 0 — Module bump

- Bump module path: `github.com/stakater/IngressMonitorController/v2` → `/v3` in `go.mod`.
- Update every import across the codebase (`gopls` / IDE rename).
- Confirm `go build ./...` still succeeds against the existing code (renamed only).
- Do **not** delete anything yet.

**Exit criteria:** clean build, all existing tests pass under the new module path.

## Phase 1 — Add CRD types (both kinds)

- **Provider spec structs (struct split — decided):**
  - `MonitorProviderCommonSpec` — provider `type`, all provider config blocks, `enableMonitorDeletion`, `creationDelay`, and `rateLimit *int` (optional, requests/minute; nil → conservative hardcoded default).
  - `ClusterMonitorProvider.spec` embeds `MonitorProviderCommonSpec` and adds `NamespaceSelector *metav1.LabelSelector` (optional; nil → referenceable from all namespaces). It has **no** `monitorNameTemplate` field — compile-time exclusion, no CEL needed; the reconciler hardcodes `{{.Namespace}}-{{.Name}}`.
  - `MonitorProvider.spec` embeds `MonitorProviderCommonSpec` and adds `MonitorNameTemplate string`.
  - `MonitorProviderStatus` — shared by both kinds. Conditions use the standard `metav1.Condition` type (gives `kubectl wait --for=condition=Ready` for free) and carry per-condition `observedGeneration`.
- Two CRD root types:
  - `ClusterMonitorProvider` with `+kubebuilder:resource:scope=Cluster`.
  - `MonitorProvider` with `+kubebuilder:resource:scope=Namespaced`.
- **`additionalPrinterColumns`** on all three CRDs via `+kubebuilder:printcolumn` markers: `*MonitorProvider` → Type, Ready, Age; `EndpointMonitor` → Ready, Age.
- Add `api/v1alpha1/endpointmonitor_types.go` changes:
  - Remove `Providers string`.
  - Add `ProviderRefs []ProviderReference` (required, min items 1, **max items 10**).
  - `ProviderReference{Kind, Name}` — **no `Namespace` field.** `Kind` is required, enum: `ClusterMonitorProvider` | `MonitorProvider`. No default.
  - Add `DeletionPolicy string` (enum `Retain` | `Delete`).
  - Expand `EndpointMonitorStatus` with `Providers []ProviderStatus` and `Conditions` (`metav1.Condition`, per-condition `observedGeneration`). `ProviderStatus` includes `Kind` and `Name` (so users can disambiguate); its `Reason` enum includes `ProviderNotFound`, `ProviderNotAllowed`, `PartialFailure`, `AuthFailed`.
- Run `make generate manifests` to produce deepcopy and CRD YAML for both kinds.
- Register both types in `groupversion_info.go`.

**Exit criteria:** `make generate manifests` produces clean output; both CRDs apply to a kind cluster; `kubectl explain clustermonitorprovider` and `kubectl explain monitorprovider` both work; `kubectl get` shows the printer columns. **CEL validation tests pass** for the bad cases: config block doesn't match `spec.type`, multiple provider blocks set on one CR, `providerRef` missing `kind`. Each rejection path has a dedicated test. Confirm the generated `ClusterMonitorProvider` schema has **no** `monitorNameTemplate` field and **does** have `namespaceSelector` (struct-split guarantees).

## Phase 2 — Provider registry and resolved config

- Add `pkg/registry/registry.go` with composite-key API: `Get/Set/Delete/Has(kind, namespace, name)`. Internally keyed by a `Key{Kind, Namespace, Name}` struct (cluster-scoped entries have empty `Namespace`).
- Each registry entry holds the constructed `MonitorServiceProxy` **plus** a per-provider rate limiter (`golang.org/x/time/rate.Limiter`) built from the provider's `rateLimit`. The limiter is shared across every `EndpointMonitor` reconcile that touches that provider — that is the reason it lives in the registry rather than being created per-reconcile.
- Add `pkg/monitors/provider-config.go` (resolved runtime config struct, tenancy-unaware). It carries `RateLimit` and the typed sub-configs (`AppInsightsConfig`, `GcloudConfig`, `GrafanaConfig`) needed by Phase 3.
- **Do not** delete `pkg/config` yet; both coexist temporarily so the build stays green.

**Exit criteria:** new packages compile in isolation. Unit tests cover:
- Concurrent Set/Get/Delete across both cluster-scoped and namespaced entries.
- Same-name collisions across kinds (cluster `foo` and `ns-a/foo`) are distinct entries.
- A registry entry's limiter is a stable instance across `Get` calls, so backpressure is genuinely shared.

## Phase 3 — Migrate `MonitorService.Setup` and rewrite the boot path

This phase touches 9 provider packages **plus** `cmd/main.go`. The boot rewrite is folded in here (not deferred to Phase 6) because deleting `SetupMonitorServicesForProviders` without replacing the `cmd/main.go` call site breaks `go build`. Keeping the build green every phase is the rule we'd otherwise violate.

- Change the interface in `pkg/monitors/monitor-service.go`:
  `Setup(p config.Provider)` → `Setup(p ProviderConfig)`.
- Change `MonitorServiceProxy.Setup` accordingly.
- Update each provider in `pkg/monitors/*`:
  - `uptimerobot/`, `pingdom/`, `pingdomtransaction/`, `statuscake/`, `uptime/`, `updown/` — type swap, same field reads.
  - `appinsights/`, `gcloud/`, `grafana/` — read from typed sub-configs. **`ProviderConfig` must carry the typed sub-configs (`AppInsightsConfig`, `GcloudConfig`, `GrafanaConfig`); define these alongside `ProviderConfig` in Phase 2** to avoid a Phase 3 type-shuffle.
- Update existing provider tests to construct `ProviderConfig`.
- Delete `SetupMonitorServicesForProviders` and the panic-on-empty path.
- Add `CreateMonitorService(pc ProviderConfig)` that returns a constructed proxy.
- **Rewrite `cmd/main.go`:**
  - Construct the registry from Phase 2.
  - **Do not register the `EndpointMonitor` reconciler in this phase.** Its v2 form depends on the `MonitorServices []*MonitorServiceProxy` field, which is being removed. Disabling it (no `SetupWithManager` call) keeps the build green and avoids zombie behavior. Phase 5 rewrites the EM reconciler against the registry and re-registers it. Between Phase 3 and Phase 5 the operator boots and is healthy but reconciles nothing — acceptable for intermediate phases.
  - Remove the `config.LoadControllerConfig` call and all `imc-config` Secret loading.
  - Delete `pkg/config/config.go` entirely and the `CONFIG_SECRET_NAME` env-var path.
  - **Add the `--cluster-resource-namespace` CLI flag** (cert-manager pattern). Default: the operator's own namespace (downward API / `OPERATOR_NAMESPACE`, as v2). This is the namespace the `ClusterMonitorProvider` reconciler will read credential Secrets from (consumed in Phase 4). Wire the flag now so the value is threaded through at manager construction.
  - Operator boots with **no providers**, no v2 Secret loader on the code path.

**Exit criteria:** `go build ./...` succeeds; `pkg/monitors/...` and `cmd/...` compile; all unit tests in `pkg/monitors/` pass; provider implementations are tenancy-unaware; `grep -ri "imc-config\|LoadControllerConfig\|CONFIG_SECRET_NAME" cmd pkg internal` returns no matches.

## Phase 4 — Provider reconcilers (both kinds)

**Two reconcilers, one shared spec-resolution helper.**

- Add `internal/controller/clustermonitorprovider_controller.go`:
  - Watches `ClusterMonitorProvider`.
  - Resolves Secrets from the **cluster-resource namespace** — the value of the `--cluster-resource-namespace` flag wired in Phase 3 (defaults to the operator namespace).
  - Writes to registry under key `{ClusterMonitorProvider, "", name}`.
  - Watches Secrets in the operator namespace; enqueues affected CRs.
  - Manages finalizer for registry cleanup on delete.
- Add `internal/controller/monitorprovider_controller.go`:
  - Watches `MonitorProvider` cluster-wide.
  - Resolves Secrets from **the CR's namespace**.
  - Writes to registry under key `{MonitorProvider, namespace, name}`.
  - Watches Secrets cluster-wide; enqueues affected CRs (only enqueues CRs whose namespace matches the changed Secret's namespace).
  - Manages finalizer for registry cleanup on delete.
- Add **shared** spec resolution: `internal/controller/providerspec_resolver.go` with a function `Resolve(ctx, client, spec, secretNamespace) (ProviderConfig, error)`. Both reconcilers call this; the only difference is the namespace they pass in. This is the dedup point.
- When writing a registry entry, construct the per-provider rate limiter from `spec.rateLimit` (nil → conservative hardcoded default, documented per provider type). On a `rateLimit` change, replace the limiter on the existing entry.
- Add RBAC kubebuilder markers:
  - `clustermonitorproviders` + status + finalizers.
  - `monitorproviders` + status + finalizers.
  - `secrets` get/list/watch **cluster-wide**.
- Regenerate `config/rbac/role.yaml`.
- **Configure label-scoped Secret cache.** In `cmd/main.go` (or wherever the manager is built), pass a `cache.Options` with a `ByObject` entry for `corev1.Secret` carrying a `LabelSelector` of `endpointmonitor.stakater.com/managed=true`. This bounds memory; RBAC stays cluster-wide. Document the requirement: every Secret referenced by a `*MonitorProvider` must carry that label, or `CredentialsResolved=False` with reason `SecretNotCached`.

**Tests:**
- Unit tests on the shared resolver (happy path, missing Secret, wrong type/block mismatch, invalid creationDelay).
- Integration tests against a kind cluster:
  - Apply ClusterMonitorProvider + Secret in operator ns → `Ready=True`, registry populated.
  - Apply MonitorProvider + Secret in team-a → `Ready=True`, registry populated.
  - Rotate the Secret → re-resolve happens automatically.
  - Delete the CR → registry entry removed.
  - Apply MonitorProvider whose Secret is in a different namespace → `CredentialsResolved=False` (Secret read scoped to CR's ns).
  - **Apply MonitorProvider whose Secret is unlabeled → `CredentialsResolved=False` with reason `SecretNotCached`. Adding the label triggers re-resolve.**
  - **Apply a ClusterMonitorProvider manifest that includes `spec.monitorNameTemplate` → the field is pruned by the API server (not in the CRD schema); the resulting monitors still use the hardcoded `{{.Namespace}}-{{.Name}}`.**
  - **Change `spec.rateLimit` on a provider → the registry entry's limiter is rebuilt with the new rate.**

**Exit criteria:** both reconcilers run; registry populates correctly for both kinds; finalizer cleanup works.

## Phase 5 — Rewrite `EndpointMonitor` reconciler

- Replace the `MonitorServices []*MonitorServiceProxy` field with `Registry *registry.Registry`.
- Reconcile loop changes:
  1. For each `providerRef`:
     - `kind: ClusterMonitorProvider` → look up `{ClusterMonitorProvider, "", ref.Name}`.
     - `kind: MonitorProvider` → look up `{MonitorProvider, em.Namespace, ref.Name}`. **Same-namespace constraint is structural** — the lookup key uses the EndpointMonitor's own namespace; there is no way to specify a different one.
  2. For a `ClusterMonitorProvider` with a non-empty `namespaceSelector`: fetch the EM's namespace and match its labels against the selector. No match → record `ProviderStatus{Ready: false, Reason: ProviderNotAllowed}`, skip this ref. (Requires `namespaces: get` RBAC — see Phase 7.)
  3. Acquire a token from the registry entry's rate limiter. If none is available → requeue the EM after the limiter's reported delay, **without mutating status** (backpressure, not failure).
  4. Iterate matched proxies, applying existing add/update logic.
  5. Aggregate per-provider status into `status.providers[]` (each entry includes `Kind` and `Name`).
  6. Set top-level `Ready` condition (`Reason=PartialFailure` if any failed).
- Add `Watches` on both `ClusterMonitorProvider` and `MonitorProvider` with map functions:
  - ClusterMonitorProvider change → enqueue every EndpointMonitor referencing that name with `kind: ClusterMonitorProvider`.
  - MonitorProvider change → enqueue every EndpointMonitor in the same namespace referencing that name with `kind: MonitorProvider`.
- Apply `deletionPolicy` override: if `spec.deletionPolicy=Retain`, skip upstream delete; if `Delete`, force it; if unset, **fall back per-provider** to that provider's `enableMonitorDeletion` (different refs on the same EM may diverge).
- **Finalizer behavior when registry entry is missing:** skip upstream cleanup, emit `OrphanedMonitor` event with the provider kind+name and the upstream monitor identifier (cached on the EM status from the last successful reconcile, if available), allow the EM to be deleted. No `Terminating` deadlock.
- **Startup window (registry warm-up):** on operator restart the registry starts empty until both provider reconcilers complete their initial list. EM reconciles that fire during that window will report `ProviderNotFound` and flip `Ready=False`, then flip back as providers re-register. **Decision: accept the flap, do not gate EM reconciles on a `RegistryReady` signal.** Rationale: a `RegistryReady` gate adds cross-reconciler coordination for a transient flap typically bounded by a few hundred milliseconds; downstream consumers should already tolerate brief `Ready=False` on operator events (restarts, leader-election handoffs). Document this in `docs/concepts.md`.

**Tests:**
- E2E on kind:
  - EndpointMonitor referencing a ClusterMonitorProvider only → works.
  - EndpointMonitor referencing a same-namespace MonitorProvider only → works.
  - EndpointMonitor referencing both kinds in one resource → both succeed independently.
  - One provider fails (e.g. wrong creds) → `Reason=PartialFailure`, other provider still applied.
  - Provider CR deleted while EM references it → next reconcile shows `ProviderNotFound`.
  - `deletionPolicy: Retain` → upstream survives EM deletion.
  - `deletionPolicy: Delete` → upstream deleted on EM deletion.
  - **Tenancy negative test:** EM in `team-a` with `kind: MonitorProvider, name: foo` where `foo` only exists in `team-b` → `ProviderNotFound`. Confirm there is no API surface that could resolve to `team-b/foo`.
  - **Finalizer fallback:** delete the `MonitorProvider` first, then the `EndpointMonitor` → EM deletion completes, `OrphanedMonitor` event present, no stuck `Terminating`.
  - **Namespace deletion:** delete the whole tenant namespace → no resources stuck `Terminating`, orphan events recorded for the surviving upstream monitors.
  - **Restart flap:** restart the operator pod while several EMs exist → confirm EMs eventually reach `Ready=True` and the flap is bounded by the time it takes both provider reconcilers to list (target: < 1s on a small kind cluster).
  - **Namespace scoping (allow):** EM in a namespace whose labels match a `ClusterMonitorProvider.spec.namespaceSelector` → provider applied.
  - **Namespace scoping (deny):** EM in a non-matching namespace → per-provider `ProviderNotAllowed`, other refs unaffected, top-level `Ready=False`.
  - **Namespace scoping (open default):** `ClusterMonitorProvider` with no `namespaceSelector` → referenceable from any namespace.
  - **Rate limit:** drive many reconciles against one provider with a low `rateLimit` → excess reconciles requeue (not fail), status does not flap to `Ready=False`, upstream call rate stays under the limit.

**Exit criteria:** end-to-end test suite passes; per-provider status is accurate; tenancy boundary cannot be crossed.

## Phase 6 — Zero-config boot verification

Most of the `cmd/main.go` rewrite landed in Phase 3 (boot path) and the reconciler registrations landed alongside each reconciler in Phase 4 / Phase 5. This phase is the final verification + cleanup pass.

- Confirm both reconcilers are registered with the manager and pick up CRs from a cold start.
- Confirm operator boots with **no providers** present (no CRs, no Secret) — runs healthy, logs "no providers configured."
- Confirm `OPERATOR_NAMESPACE` resolution still works (downward API + env override) and that `--cluster-resource-namespace` correctly overrides it when set; the `ClusterMonitorProvider` reconciler reads Secrets from the resolved namespace.
- **Fix OpenShift detection at process `init()`.** Today `cmd/main.go`'s `init()` calls `kube.IsOpenshift()`, which builds a client and hits the API server — a brief API outage at pod start crash-loops the operator, contradicting the "boot healthy with zero config" goal. Move detection to a lazy, cached check evaluated inside the route-aware code path (URL resolution) on first use. Operator boot must not depend on cluster reachability.
- Sweep for any remaining v2 vestiges: stray references to `imc-config`, comments mentioning the old loader, dead test fixtures.

**Exit criteria:** fresh deployment with no provider CRs produces a healthy operator pod; `kubectl logs` shows the expected "ready, awaiting providers" message; `grep -ri "imc-config\|CONFIG_SECRET_NAME" .` returns only docs/migration mentions.

## Phase 7 — Helm chart, RBAC, and manifests

- Update `charts/ingressmonitorcontroller/`:
  - Remove the `imc-config` Secret rendering and the values block that fed it.
  - Add CRD installation for both `ClusterMonitorProvider` and `MonitorProvider`.
  - Update `values.yaml` — remove `config:`; document any new chart options.
  - Update `NOTES.txt` with a clear "v3 breaking change" warning and link to the migration doc.
- Update `config/crd/bases/`, `config/rbac/`, `config/samples/`:
  - Sample for ClusterMonitorProvider (UptimeRobot).
  - Sample for MonitorProvider (UptimeRobot).
  - Sample EndpointMonitor mixing both kinds.
- Bundle: regenerate `bundle/manifests/`.
- ClusterRole verbs:
  - `endpointmonitor.stakater.com/{endpointmonitors,monitorproviders,clustermonitorproviders}` — full + status + finalizers.
  - `core/secrets` — get/list/watch cluster-wide.
  - `core/namespaces` — get/list/watch (**required** — the `EndpointMonitor` reconciler reads namespace labels to evaluate `ClusterMonitorProvider.spec.namespaceSelector`).
  - `networking.k8s.io/ingresses`, `route.openshift.io/routes` — as today.
- **Write `docs/migration-v2-to-v3.md` in this phase** (not Phase 8) so `NOTES.txt` has a target to link to:
  - Operational upgrade sequence: (1) delete v2 `EndpointMonitor`s, (2) uninstall v2 operator, (3) install v3 CRDs, (4) install v3 operator, (5) create `*MonitorProvider` CRs + labeled credential Secrets, (6) reapply EMs in v3 shape. **Step ordering matters** — installing v3 CRDs over v2 EMs causes schema-rejection errors on those EMs until they are recreated.
  - Single-tenant case: how a v2 `imc-config` with one or two providers maps to `ClusterMonitorProvider`s in the operator namespace.
  - Multi-tenant case: deciding shared vs. local per provider; moving credentials into per-tenant Secrets.
  - For each provider type (UptimeRobot, Pingdom, PingdomTransaction, StatusCake, Uptime, Updown, AppInsights, GCloud, Grafana): side-by-side v2 Secret YAML and v3 CR + Secret pair, including the `endpointmonitor.stakater.com/managed=true` label.

**Exit criteria:** `helm install` on a fresh cluster succeeds without provider config; `helm upgrade` from v2 surfaces a clear breaking-change warning; migration doc covers every supported provider and the upgrade sequence.

## Phase 8 — Concept docs and examples

The migration doc landed in Phase 7. This phase covers the remaining user-facing documentation.

- Add `docs/concepts.md` — covers tenancy model, `ClusterMonitorProvider` vs `MonitorProvider`, when to use which, the registry, the same-namespace constraint, the operator-restart flap (Phase 5), the `endpointmonitor.stakater.com/managed=true` Secret label requirement, and the per-provider rate limiter (`spec.rateLimit`).
- Add `docs/monitorprovider.md` — fields, examples per provider type, both kinds shown side by side.
- Add `docs/multi-tenancy.md` — explicit walkthrough of the platform-team + tenant-team setup. Covers RBAC patterns for tenant admins and `ClusterMonitorProvider.spec.namespaceSelector` for restricting a shared provider to specific tenant namespaces. Cross-link to `migration-v2-to-v3.md` for the multi-tenant cutover case.
- Rewrite `README.md` quickstart to use `MonitorProvider`.
- Update each provider's existing config doc to show the new shape.
- Update `examples/configs/` (delete v2-style YAMLs, add v3 CR samples with labeled Secrets) and `examples/endpointMonitor/`.

**Exit criteria:** a new user can go from zero to a working multi-tenant setup by reading only the README, `concepts.md`, and `multi-tenancy.md`.

## Phase 9 — Release

- **Regenerate the OLM bundle:** `make bundle` (the v2 bundle in `bundle/manifests/` is stale — new CRDs, new RBAC markers, new sample CRs). Confirm `operator-sdk bundle validate ./bundle` passes.
- Tag `v3.0.0-rc.1`.
- Cut a pre-release in GitHub with `docs/migration-v2-to-v3.md` front and center.
- **Release notes must state the E2E coverage limitation explicitly:** only UptimeRobot and StatusCake are validated end-to-end against live upstream APIs; the other seven providers are mock-tested only (no test accounts available). Users of the mock-only providers should validate in a staging environment before production.
- After bake time, tag `v3.0.0`.

## Open questions to confirm before starting

These were settled during design discussion but flagged here in case implementation surfaces nuance:

1. **Partial-failure behavior on EndpointMonitor:** confirmed degraded-but-running. Top-level `Ready=False`, other providers continue.
2. **`deletionPolicy` fallback when unset:** confirmed fall back to `MonitorProvider.spec.enableMonitorDeletion`.
3. **Cross-kind name collisions:** confirmed allowed and disambiguated by `kind`.
4. **Operator RBAC for Secrets:** confirmed cluster-wide.

Will revisit during Phase 5 if the EndpointMonitor reconciler surfaces edge cases.

## Phase complexity estimates

| Phase | Complexity | Risk |
|---|---|---|
| 0 — Module bump | S | None |
| 1 — CRD types | M | CEL validation; cluster-vs-namespaced spec divergence (`namespaceSelector` cluster-only, `monitorNameTemplate` namespaced-only) |
| 2 — Registry + ProviderConfig | S | None |
| 3 — Migrate `Setup` signatures + boot rewrite | **L** | 9 provider packages + `cmd/main.go` rewrite + `pkg/config` delete in one phase; risk is missing a less-used provider's typed sub-config |
| 4 — Two provider reconcilers | L | Finalizer + label-scoped Secret cache config; cross-namespace Secret list/watch perf |
| 5 — EndpointMonitor reconciler rewrite | L | Status aggregation, per-ref `deletionPolicy` precedence, finalizer fallback, restart-flap, `namespaceSelector` + rate-limiter wiring |
| 6 — Zero-config boot verification | S | Low — OpenShift `init()` → lazy-detection refactor is the only code change |
| 7 — Helm/RBAC + migration doc | M | NOTES.txt + operational upgrade sequence |
| 8 — Concept docs and examples | M | Multi-tenancy concepts doc requires care |
| 9 — Release + bundle regen | S | Communications + bundle validation |

Total: roughly 6 phases of substantive engineering work + 2 of plumbing + 1 of release. Realistic for a 6–8 week effort with one senior engineer; longer with reviews and CI ramp-up.
