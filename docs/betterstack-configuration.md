# Better Stack Configuration

## Getting an API token

Better Stack's API uses a bearer token, created under **Settings → API tokens**
in the Uptime dashboard. The token is account-wide: it can create, edit and
delete every monitor on the account.

Put it in the controller's config secret as `apiToken`:

```yaml
providers:
  - name: BetterStack
    apiToken: <token>
```

`apiKey` is accepted as an alias so a config written for another provider still
works, and `apiURL` defaults to `https://uptime.betterstack.com`.

`alertContacts` sets the **default escalation policy** for every monitor this
controller creates:

```yaml
providers:
  - name: BetterStack
    apiToken: <token>
    alertContacts: "12345"   # escalation policy id
```

Better Stack has no notion of a contact list attached to a monitor — who gets
notified is decided by the escalation policy — so the provider-neutral
`alertContacts` setting maps onto `policy_id`. A CR naming its own `policyID`
overrides it.

Without either, Better Stack applies its own default: email to the account
owner. Monitors alert out of the box; a policy is how you route them somewhere
else.

**Escalation policies are a paid feature.** On a free account, creating one
returns `403 Cannot create escalation policy. Please upgrade your account`, so
`alertContacts` and `policyID` have nothing to point at. The per-monitor
`email`, `sms`, `call` and `push` booleans work on every plan and are the way to
vary alerting there. Slack, when connected, is configured account-wide in Better
Stack rather than per monitor — it applies to everything this controller
creates.

## Configuration

Per-monitor settings go under `betterStackConfig` on an `EndpointMonitor`:

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: web
spec:
  forceHttps: true
  healthEndpoint: /api/health
  urlFrom:
    ingressRef:
      name: web
  betterStackConfig:
    checkFrequency: 60
    sms: "true"
    call: "true"
    policyID: "12345"
```

| Field | Description |
|---|---|
| `checkFrequency` | Seconds between checks. Better Stack's default (180) applies when unset. |
| `monitorType` | `status`, `expected_status_code`, `keyword`, `keyword_absence`, `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap`, `dns`. Defaults to `status`. |
| `expectedStatusCodes` | Comma separated, e.g. `"200,302"`. Only used with `expected_status_code`. |
| `requiredKeyword` | Keyword that must be present or absent. Only used with the keyword types. |
| `policyID` | Escalation policy to attach. Without one, Better Stack applies the team default. |
| `regions` | Comma separated, e.g. `"eu,us"`. |
| `requestTimeout` | Seconds before a check counts as failed. |
| `confirmationPeriod` | Seconds a failure must persist before the monitor is marked down. |
| `recoveryPeriod` | Seconds of health before recovery. **Must be one of 0, 60, 180, 300, 900, 1800, 3600, 7200** — Better Stack rejects anything else, so an invalid value is dropped with a log rather than failing the whole monitor. |
| `teamWait` | Seconds before escalating to the rest of the team. |
| `verifySSL`, `followRedirects`, `rememberCookies`, `paused`, `email`, `sms`, `call`, `push` | `"true"` or `"false"`. |

### Booleans are strings

They are `"true"` / `"false"` rather than real booleans so that **unset is
distinct from false**. An unset field is omitted from the API request entirely,
which leaves Better Stack's own default — or a change someone made in their UI —
untouched. A real `bool` could not express that: Go's zero value would send
`false` and silently overwrite it.

### Monitoring a redirect

Better Stack refuses a monitor that expects a 3xx status while following
redirects or keeping cookies:

```
Cannot follow redirects when expecting a 3xx status code
Cannot keep cookies when redirecting when expecting a 3xx status code
```

Probing a redirect is a legitimate check — an apex that 302s to the app is worth
watching in its own right — so when `expectedStatusCodes` contains a 3xx and
neither field is set explicitly, both are sent as `false`:

```yaml
  betterStackConfig:
    monitorType: expected_status_code
    expectedStatusCodes: "301,302"
```

Setting either one explicitly overrides that, and Better Stack will reject the
request.

## Testing against the live API

The provider has integration tests behind a build tag. They create a paused
monitor against `https://example.com/`, assert the attributes round-trip, and
delete it — including on failure.

```bash
BETTERSTACK_API_TOKEN=<token> go test -tags=integration ./pkg/monitors/betterstack/...
```

Without the token they skip.
