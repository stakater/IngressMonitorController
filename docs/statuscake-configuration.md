# StatusCake Configuration

## Basic
The following properties need to be configured for Statuscake, in addition to the general properties listed 
 in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                         |
|----------|-----------------------------------------------------|
| username | Account username for authentication with Statuscake |

## Advanced

Currently additional Statuscake configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| statuscake.monitor.stakater.com/check-rate               | Set Check Rate for the monitor (default: 300)    |
| statuscake.monitor.stakater.com/test-type                | Set Test type - HTTP, TCP, PING (default: HTTP)  |
| statuscake.monitor.stakater.com/paused                   | Pause the service                                |
| statuscake.monitor.stakater.com/ping-url                 | Webhook for alerts                               |
| statuscake.monitor.stakater.com/follow-redirect          | Enable ingress redirects                         |
| statuscake.monitor.stakater.com/port                     | TCP Port                                         |
| statuscake.monitor.stakater.com/trigger-rate             | Minutes to wait before sending an alert          |
| statuscake.monitor.stakater.com/contact-group            | Contact Group to be alerted.                     |
| statuscake.monitor.stakater.com/basic-auth-user          | Required for [basic-authenticationchecks](#basic-auth-checks)  |


### Basic Auth checks

Statuscake supports checks completing basic auth requirements. The annotation `statuscake.monitor.stakater.com/basic-auth-user` can be used to trigger the Ingress Monitor attempting to configure this setting. The value of the annotation should be the *username* to be configured. The Ingress Monitor Controller will then attempt to access an OS env variable of the same name which will return the *password* that should be used. The env variable can be mounted within the Ingress Monitor Controller container via a secret.

For example; the annotation `statuscake.monitor.stakater.com/basic-auth-user: 'my-service-username'` will set the username field to the value `my-service-username` and will retrieve the password via `os.Getenv('my-service-username')` and set this appropriately. If the password is not found/set the annotation will be skipped.