# Pingdom Configuration

## Note
Currently we do not have access to Pingdom account that's why Tests are not verified. Community members having Pingdom account are welcome to contribute in Test Cases.

## Basic
The following properties need to be configured for Pingdom, in addition to the general properties listed 
in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                      |
|----------|--------------------------------------------------|
| username | Account username for authentication with Pingdom |
| password | Account password for authentication with Pingdom |

## Optional
The following optional property can be included for Pingdom accounts which require multi-user authentication.
More information can be found [Here](https://www.pingdom.com/api/2.1/#multi-user+authentication)

| Key               | Description                                              |
|-------------------|----------------------------------------------------------|
| accountEmail      | Email account for multi-user authentication with Pingdom |
| alertIntegrations | Comma separated list of integration ids                  |

## Advanced

Currently additional pingdom configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| pingdom.monitor.stakater.com/resolution                  | The pingdom check interval in minutes            |
| pingdom.monitor.stakater.com/send-notification-when-down | How many failed check attempts before notifying  |
| pingdom.monitor.stakater.com/paused                      | Set to "true" to pause checks                    |
| pingdom.monitor.stakater.com/notify-when-back-up         | Set to "false" to disable recovery notifications |
| pingdom.monitor.stakater.com/request-headers             | Custom pingdom request headers (e.g. {"Accept"="application/json"}) |
| pingdom.monitor.stakater.com/basic-auth-user             | Required for basic-authentication checks - [see below](#basic-auth-checks) |
| pingdom.monitor.stakater.com/should-contain              | Set to text string that has to be present in the HTML code of the page (configures "Should contain") |
| pingdom.monitor.stakater.com/tags                        | Comma separated set of tags to apply to check (e.g. "testing,aws") |
| pingdom.monitor.stakater.com/alert-integrations                | Comma separated set list of integrations ids (e.g. "91166,12168") |

### Basic Auth checks

Pingdom supports checks completing basic auth requirements. The annotation `pingdom.monitor.stakater.com/basic-auth-user` can be used to trigger the Ingress Monitor attempting to configure this setting. The value of the anotation should be the username to be configured. The Ingress Monitor Controller will then attempt to access an OS env variable of the same name which will return the password that should be used. The env variable can be mounted within the Ingress Monitor Controller container via a secret.

For example; the annotation `pingdom.monitor.stakater.com/basic-auth-user: 'health-service'` will set the username field to 'health-service' and will retrieve the password via `os.Getenv('health-service')` and set this appropriately.
