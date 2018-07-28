# StatusCake Configuration

## Basic
The following properties need to be configured for Statuscake, in addition to the general properties listed 
 in the [Configuration section of the README](../README.md#configuration):

| Key      | Description                                         |
|----------|-----------------------------------------------------|
| username | Account username for authentication with Statuscake |
| password | Account password for authentication with Statuscake |

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