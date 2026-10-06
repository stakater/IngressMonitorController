# Uptime Kuma Configuration

Uptime Kuma support speaks the Uptime Kuma **2.x** socket.io API via [go-uptime-kuma-client](https://github.com/breml/go-uptime-kuma-client). Older Uptime Kuma versions (1.x) are not supported.

Notes:

- Uptime Kuma API keys do **not** work for creating, updating or deleting monitors, so the controller authenticates with **username and password**.
- **2FA must be disabled** for the account the controller uses — create a dedicated service account in Uptime Kuma without two-factor authentication.
- Uptime Kuma allows only 20 logins per minute for all clients together. The controller therefore keeps one connection for its lifetime and backs off for up to 5 minutes after a failed login, so a wrong password cannot lock out the Uptime Kuma UI.
- Point `apiURL` at the internal HTTP service (e.g. `http://uptime-kuma.uptime-kuma.svc:3001`) rather than an ingress to avoid self-signed certificate issues.

## Configuration

Add a provider entry to the controller's config:

```yaml
providers:
  - name: UptimeKuma
    apiURL: http://uptime-kuma.uptime-kuma.svc:3001
    username: <your-username>
    password: <your-password>
```

Additional uptime kuma configurations can be added to the EndpointMonitor through these fields:

| Fields         | Description                                                                                              |
|:--------------:|:--------------------------------------------------------------------------------------------------------:|
| Interval       | The uptime kuma check interval in seconds (minimum 20, defaults to 60)                                    |
| MonitorType    | The uptime kuma monitor type (http or keyword)                                                            |
| KeywordExists  | `yes` keeps the monitor up while the keyword is present in the response, `no` keeps it up while the keyword is absent (Only if monitor-type is keyword) |
| KeywordValue   | keyword to check on URL (e.g.'search' or '404') (Only if monitor-type is keyword)                         |
| Notifications  | Comma-separated uptime kuma notification IDs to attach to this monitor                                    |

### Keyword monitors

A keyword monitor compares the response body against `keywordValue`:

- `keywordExists: "yes"` (the default) — the monitor is **up** while the keyword is in the response and goes **down** when it disappears, for example `keywordValue: "search"` for a search page that is expected to render.
- `keywordExists: "no"` — the monitor is **up** only while the keyword is *absent*, so it goes **down** (and alerts) as soon as the keyword shows up, for example `keywordValue: "404"` to be alerted when the endpoint starts answering with a 404 page.

This matches the `UptimeRobot` provider, where `keywordExists` selects between `keyword_type` 1 (contains) and 2 (does not contain).

> **Note**
> `keywordExists` and `keywordValue` are strings. Quote them, otherwise YAML turns `no` into a boolean and `404` into a number and the API server rejects the resource.

### Checking fields

The controller sets the remaining Uptime Kuma check fields to the values the Kuma UI uses by default, since Uptime Kuma has no default for them and rejects a monitor without them:

| Field           | Value                     |
|:---------------:|:--------------------------|
| Retry interval  | the same as the interval  |
| Max redirects   | 10                        |
| Timeout         | 80% of the interval       |
| Accepted status | 200-299                   |

### Fetching notification IDs from Uptime Kuma

Notification IDs are the identifiers of notification targets (e.g. a Slack or ntfy integration) configured in Uptime Kuma. You can find the ID of a notification in **Settings > Notifications** — it is shown in the URL when editing a notification, or in the notification list API.

The controller checks the configured IDs against the notifications of the account it logs in with and refuses to create or update a monitor if one of them does not exist. Uptime Kuma itself accepts such a monitor but silently drops the notifications, which made the controller retry a failing update on every reconcile.

## Example

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHttps: true
  url: https://stakater.com/
  uptimeKumaConfig:
    interval: 120
    monitorType: keyword
    keywordExists: "no"
    keywordValue: "404"
    notifications: "1,2"
```
