# Updown Configuration

## Note

| Caveats    | Description                                      |
|----------|-----------------------------------------------------|
| Email alert addresses   | Currently, email addresses for check alerts can be assigned to a check, neither by its go client nor API. The only way to do it is by manually adding them on their website. A request has been added on updown's feature addition forum [link](https://updown.uservoice.com/forums/177972-general/suggestions/37334926-crud-for-setting-email-and-phone-alerts-for-a-chec)  |




## Compulsory Configuration

The following properties need to be configured for Updown, in addition to the general properties listed 
in the [Configuration section of the README](../README.md#configuration):


| Key      | Description                                      |
|----------|--------------------------------------------------|
| apiKey   | API key of an account                            |


## Additional Configuration

Additional updown configurations can be added through these fields:

|                        Fields                       |                    Description                   |
|----------------------------------------------------------|--------------------------------------------------|
| Enable  | Set to "false" to disable checks                 |
| Period                       | The pingdom check interval in seconds, it accepts `only` these values: 15, 30, 60, 120, 300, 600, 1800, 3600  |
| PublishPage | Status page be public or not ("true" or "false")|
| RequestHeaders              | Custom updown request headers (e.g. {"Accept"="application/json"}) |


## Example: 

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHtpps: true
  url: https://stakater.com/
  updownConfig:
    enable: true
    period: 15
    publishPage: true
    requestHeaders: {"Accept"="application/json"}
```