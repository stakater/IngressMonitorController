# StatusCake Configuration

## Basic
The following properties need to be configured for Statuscake, in addition to the general properties listed 
 in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                         |
|----------|-----------------------------------------------------|
| username | Account username for authentication with Statuscake |

## Advanced

Currently additional Statuscake configurations can be added through these fields:

|                        Fields                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| CheckRate               | Set Check Rate for the monitor (default: 300)    |
| TestType                | Set Test type - HTTP, TCP, PING (default: HTTP)  |
| Paused                   | Pause the service                                |
| PingURL                 | Webhook for alerts                               |
| FollowRedirect          | Enable ingress redirects                         |
| Port                     | TCP Port                                         |
| TriggerRate             | Minutes to wait before sending an alert          |
| ContactGroup            | Contact Group to be alerted.                     |
| TestTags                | Comma separated list of tags                     |
| BasicAuthUser          | Required for [basic-authenticationchecks](#basic-auth-checks)  |


### Basic Auth checks

Statuscake supports checks completing basic auth requirements. In `EndpointMonitor` the field `basicAuthUser` can be used to trigger the Ingress Monitor attempting to configure this setting. The value of the field should be the *username* to be configured. The Ingress Monitor Controller will then attempt to access an OS env variable of the same name which will return the *password* that should be used. The env variable can be mounted within the Ingress Monitor Controller container via a secret.

For example; setting the field like `basic-auth-user: 'my-service-username'` will set the username field to the value `my-service-username` and will retrieve the password via `os.Getenv('my-service-username')` and set this appropriately. 

## Example: 

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHttps: true
  url: https://stakater.com/
  statusCakeConfig:
    basicAuthUser: my-service-username
    checkRate: 300
    testType: HTTP
    paused: false
```
