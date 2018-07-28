# Pingdom Configuration

## Basic
The following properties need to be configured for Pingdom, in addition to the general properties listed 
in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                      |
|----------|--------------------------------------------------|
| username | Account username for authentication with Pingdom |

## Advanced

Currently additional pingdom configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| monitor.stakater.com/pingdom/resolution                  | The pingdom check interval in minutes            |
| monitor.stakater.com/pingdom/send-notification-when-down | How many failed check attempts before notifying  |
| monitor.stakater.com/pingdom/paused                      | Set to "true" to pause checks                    |
| monitor.stakater.com/pingdom/notify-when-back-up         | Set to "false" to disable recovery notifications |