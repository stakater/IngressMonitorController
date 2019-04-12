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

Additional updown configurations can be added to each ingress object using annotations, current supported annotations are:

|                        Annotations                       |                    Description                   |
|----------------------------------------------------------|--------------------------------------------------|
| updown.monitor.stakater.com/enable  | Set to "false" to disable checks                 |
| updown.monitor.stakater.com/period                       | The pingdom check interval in seconds, it accepts `only` these values: 15, 30, 60, 120, 300, 600, 1800, 3600  |
| updown.monitor.stakater.com/enable | Check be enabled or not ("true" or "false") |
| updown.monitor.stakater.com/publish-page | Status page be public or not ("true" or "false")|
| updown.monitor.stakater.com/request-headers              | Custom updown request headers (e.g. {"Accept"="application/json"}) |