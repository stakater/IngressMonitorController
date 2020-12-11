# Uptime Configuration
## Fetching alert contacts from UpTime

In order to use Ingress Monitor controller, you need to have contacts added to your account in form of groups. Once you add them via Dashboard, you will need just the group name. You can add as many groups as you want with a `,` separator in between. (A `Default` group is established upon sign-up)

## Configuration

Additional uptime configurations can be added through following fields:

|                        Fields                    |                    Description                               |
|:----------------------------------------------------:|:------------------------------------------------------------:|
| Interval            | The uptime check interval in seconds                    |
| CheckType        | The uptime check type that can be HTTP/DNS/ICMP etc. |
| Contacts | Add one or more contact groups separated by `,` |
| Locations | Add different locations for the check |

## Example: 

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHttps: true
  url: https://stakater.com/
  uptimeConfig:
    interval: 60
    checkType: HTTP
    contacts: "133,132"
    locations: "sea,fr"
    tags: "core" # ignored if not passed
```
