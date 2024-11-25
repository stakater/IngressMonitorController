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
| Paused                  | Pause the service                                |
| PingURL                 | Webhook for alerts                               |
| FollowRedirect          | Enable ingress redirects                         |
| Port                    | TCP Port                                         |
| TriggerRate             | Minutes to wait before sending an alert          |
| ContactGroup            | Contact Group to be alerted.                     |
| TestTags                | Comma separated list of tags                     |
| FindString              | String to look for within the response           |
| BasicAuthUser           | Required for [basic-authenticationchecks](#basic-auth-checks)  |
| BasicAuthSecret         | Allows for an alternate method of adding basic-auth to checks |
| Regions                 | Regions to execute the check from                |
| RawPostData             | Add data to change the request to a POST         |
| UserAgent               | Add a user agent string to the request           |


### Basic Auth checks

Statuscake supports checks completing basic auth requirements. In `EndpointMonitor` the field `basicAuthUser` can be used to trigger the Ingress Monitor attempting to configure this setting. The value of the field should be the *username* to be configured. The Ingress Monitor Controller will then attempt to access an OS env variable of the same name which will return the *password* that should be used. The env variable can be mounted within the Ingress Monitor Controller container via a secret.

For example; setting the field like `basic-auth-user: 'my-service-username'` will set the username field to the value `my-service-username` and will retrieve the password via `os.Getenv('my-service-username')` and set this appropriately.

In addition to the previous method, you can use the `basicAuthSecret` field to define a secret that should be read by the monitor which contains the basic-auth data. This secret should only contain the keys `username` and `password`. It expects the values for those keys to be strings. NOT base64 encoded strings. Furthermore, the secret must be present in the same namespace as the IngressMonitorController operator. This ensures that we can keep the permissions of the operator to be as small as possible.

So for example, if you have a secret called `my-deployment-secret` it should contain the data `username: my-user` and `password: MyPassword1!` and you should set to `basicAuthSecret: my-deployment-secret`. This will ensure that the monitor can read the basic-auth data correctly.

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
