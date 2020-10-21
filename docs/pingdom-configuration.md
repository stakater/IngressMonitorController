# Pingdom Configuration

## Note
Currently we do not have access to Pingdom account that's why Tests are not verified. Community members having Pingdom account are welcome to contribute in Test Cases.

## Basic
The following properties need to be configured for Pingdom, in addition to the general properties listed
in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                      |
|----------|--------------------------------------------------|
| apiKey | Pingdom API Token generated inside My Pingdom |

## Optional
The following optional property can be included if you want to declare some default options, without re-declaring them for each EndpointMonitor.
You are able to override any of them via EndpointMonitor specific options.

| Key               | Description                                              |
|-------------------|----------------------------------------------------------|
| alertIntegrations | `-` separated list of integration ids                  |
| teamAlertContacts | `-` separated list of teams ids                  |
| alertContacts     | `-` separated list of alert contacts ids                  |
## Advanced

Currently additional pingdom configurations can be added through these fields:

|                        Fields                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| Resolution                  | The pingdom check interval in minutes            |
| SendNotificationWhenDown | How many failed check attempts before notifying  |
| Paused                      | Set to "true" to pause checks                    |
| NotifyWhenBackUp         | Set to "false" to disable recovery notifications |
| RequestHeaders             | Custom pingdom request headers (e.g. {"Accept"="application/json"}) |
| BasicAuthUser             | Required for basic-authentication checks - [see below](#basic-auth-checks) |
| ShouldContain              | Set to text string that has to be present in the HTML code of the page (configures "Should contain") |
| Tags                        | Comma separated set of tags to apply to check (e.g. "testing,aws") |
| AlertIntegrations                | `-` separated set list of integrations ids (e.g. "91166-12168") |
| AlertContacts                | `-` separated contact id's (e.g. "1234567_8_9-9876543_2_1") to override the [default alertContacts](https://github.com/stakater/IngressMonitorController/blob/master/README.md#usage)|
| TeamAlertContacts            | Teams to alert.  `-` separated set list of teams ids (e.g. "1234567_8_9-9876543_2_1)|

### Basic Auth checks

Pingdom supports checks completing basic auth requirements. In `EndpointMonitor` the field `basicAuthUser` can be used to trigger the Ingress Monitor attempting to configure this setting. The value of the field should be the username to be configured. The Ingress Monitor Controller will then attempt to access an OS env variable of the same name which will return the password that should be used. The env variable can be mounted within the Ingress Monitor Controller container via a secret.

For example; setting the field like `basicAuthUser: health-service` will set the username field to 'health-service' and will retrieve the password via `os.Getenv('health-service')` and set this appropriately.


## Example: 

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHttps: true
  url: https://stakater.com/
  pingdomConfig:
    resolution: 10
    sendNotificationWhenDown: true
    paused: false
    notifyWhenBackUp: false
    requestHeaders: {"Accept"="application/json"}
    basicAuthUser: health-service
    shouldContain: "must have text"
    tags: "testing,aws"
    alertIntegrations: "91166-12168"
    alertContacts: "1234567_8_9-9876543_2_1,1234567_8_9-9876543_2_2"
    teamAlertContacts: "1234567_8_9-9876543_2_1,1234567_8_9-9876543_2_2"
```
