# Pingdom Configuration

Currently additional pingdom configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| monitor.stakater.com/pingdom/resolution                  | The pingdom check interval in minutes            |
| monitor.stakater.com/pingdom/send-notification-when-down | How many failed check attempts before notifying  |
| monitor.stakater.com/pingdom/paused                      | Set to "true" to pause checks                    |
| monitor.stakater.com/pingdom/notify-when-back-up         | Set to "false" to disable recovery notifications |