# Uptime Kuma Configuration

Uptime Kuma support speaks the Uptime Kuma **2.x** socket.io API via [go-uptime-kuma-client](https://github.com/breml/go-uptime-kuma-client). Older Uptime Kuma versions (1.x) are not supported.

Notes:

- Uptime Kuma API keys do **not** work for creating, updating or deleting monitors, so the controller authenticates with **username and password**.
- **2FA must be disabled** for the account the controller uses — create a dedicated service account in Uptime Kuma without two-factor authentication.
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
| KeywordExists  | Alert if value exist (yes) or doesn't exist (no) (Only if monitor-type is keyword)                        |
| KeywordValue   | keyword to check on URL (e.g.'search' or '404') (Only if monitor-type is keyword)                         |
| Notifications  | Comma-separated uptime kuma notification IDs to attach to this monitor                                    |

### Fetching notification IDs from Uptime Kuma

Notification IDs are the identifiers of notification targets (e.g. a Slack or ntfy integration) configured in Uptime Kuma. You can find the ID of a notification in **Settings > Notifications** — it is shown in the URL when editing a notification, or in the notification list API.

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
    keywordExists: no
    keywordValue: 404
    notifications: "1,2"
```
