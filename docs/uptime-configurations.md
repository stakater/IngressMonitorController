# Uptime Configuration
## Fetching alert contacts from UpTime

In order to use Ingress Monitor controller, you need to have contacts added to your account in form of groups. Once you add them via Dashboard, you will need just the group name. You can add as many groups as you want in annotations with a `,` separator in between. (A `Default` group is established upon sign-up)

## Annotations

Additional uptime configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                    |                    Description                               |
|:----------------------------------------------------:|:------------------------------------------------------------:|
| uptime.monitor.stakater.com/interval            | The uptime check interval in seconds                    |
| uptime.monitor.stakater.com/check_type        | The uptime check type that can be HTTP/DNS/ICMP etc. |
| uptime.monitor.stakater.com/contacts | Add one or more contact groups separated by `,` |
| uptime.monitor.stakater.com/locations | Add different locations for the check |
