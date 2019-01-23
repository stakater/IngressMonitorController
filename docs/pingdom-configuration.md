# Pingdom Configuration

## Basic
The following properties need to be configured for Pingdom, in addition to the general properties listed 
in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                      |
|----------|--------------------------------------------------|
| username | Account username for authentication with Pingdom |
| password | Account password for authentication with Pingdom |

## Advanced

Currently additional pingdom configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| pingdom.monitor.stakater.com/resolution                  | The pingdom check interval in minutes            |
| pingdom.monitor.stakater.com/send-notification-when-down | How many failed check attempts before notifying  |
| pingdom.monitor.stakater.com/paused                      | Set to "true" to pause checks                    |
| pingdom.monitor.stakater.com/notify-when-back-up         | Set to "false" to disable recovery notifications |
| pingdom.monitor.stakater.com/request-headers             | Custom pingdom header (e.g {"Accept"="application/json"}) |