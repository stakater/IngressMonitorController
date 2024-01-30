# Pingdom Transaction Configuration

This document describes how to configure a Pingdom Transaction monitor using the `PingdomTransactionConfig` struct in the EndpointMonitor custom resource.

## `PingdomTransactionConfig` Structure

The `PingdomTransactionConfig` struct is defined as follows:

| Field | Type | Description |
|-------|------|-------------|
| Paused | bool | Check status: active or inactive |
| CustomMessage | string | Custom message that is part of the email and webhook alerts |
| Interval | int | TMS test intervals in minutes. Allowed intervals: 5,10,20,60,720,1440 |
| Region | string | Name of the region where the check is executed. Supported regions: us-east, us-west, eu, au |
| SendNotificationWhenDown | int64 | Send notification when down X times |
| SeverityLevel | string | Check importance- how important are the alerts when the check fails. Allowed values: low, high |
| Steps | []PingdomStep | Steps to be executed as part of the check |
| Tags | []string | List of tags for a check |
| AlertIntegrations | string | `-` separated set list of integrations ids |
| AlertContacts | string | `-` separated contact id's |
| TeamAlertContacts | string | `-` separated team id's |

Each `PingdomStep` is defined as follows:

| Field | Type | Description |
|-------|------|-------------|
| Args | map[string]string | Contains the HTML element with assigned value |
| Function | string | Contains the function that is executed as part of the step |

## Configuration

To configure a Pingdom Transaction monitor, you need to specify the fields in the `PingdomTransactionConfig` section of your EndpointMonitor custom resource.

Here's an example:

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: manual-pingdom-transaction-check
spec:
  # url is not used, but required for ingressmonitorcontroller to work
  url: https://www.google.com
  pingdomTransactionConfig:
    steps:
      - function: go_to
        args:
          url: https://www.google.com
      - function: fill
        args:
          input: textarea[name=q]
          value: kubernetes
      - function: submit
        args:
          form: form
      - function: exists
        args:
          element: '#rso'
    alertContacts: "14901866"
    alertIntegrations: "132625"
    #teamAlertContacts: "14901866"
    custom_message: "This is a custom message"
    paused: false
    region: us-east
    interval: 60
    send_notification_when_down: 3
    severity_level: low
    tags:
      - testing
      - manual
```

In this example, we run a transaction check against Google. The check will fail if the search results page does not contain the search term "kubernetes".

For full details on the available functions and arguments, please refer to the [Pingdom Transaction Checks API documentation](https://docs.pingdom.com/api/#section/TMS-Steps-Vocabulary/Script-transaction-checks).

Please refer to the [Pingdom Transaction Checks API documentation](https://docs.pingdom.com/api/#operation/getAllCheckss) for more information on how to configure transaction checks.

## Using Passwords and Secrets

You may have fields that contain secret values, such as passwords. To avoid storing these values in plain text, you can use a template function to reference a secret value. This makes sure that the secret value is not stored in plain text in your EndpointMonitor custom resource.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret-name
stringData:
  my-secret-key: kubernetes operator
  admin-password: adminPass
---
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: manual-pingdom-transaction-check-with-password
spec:
  url: https://www.google.com
  pingdomTransactionConfig:
    steps:
      - function: go_to
        args:
          url: https://www.google.com
      - function: fill
        args:
          input: textarea[name=q]
          # this will replaced with the value of the secret before sending the request to pingdom
          value: '{secret:my-secret-name:my-secret-key}'
      - function: basic_auth
        args:
          user: admin
          password: '{secret:my-secret-name:admin-password}'
      - function: submit
        args:
          form: form
      - function: exists
        args:
          element: '#rso'
```

The secret must be located in the same namespace as the IngressMonitorController. The operator can only retrieve secrets from his own namespace. The template function is only activated for the `value` and `password` fields.
