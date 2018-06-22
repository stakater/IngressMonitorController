# StatusCake Configuration

Currently additional Statuscake configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation                        |                    Description                   |
|:--------------------------------------------------------:|:------------------------------------------------:|
| monitor.stakater.com/statuscake-check-rate               | Set Check Rate for the monitor (default: 300)    |
| monitor.stakater.com/statuscake-test-type                | Set Test type - HTTP, TCP, PING (default: HTTP)  |
| monitor.stakater.com/statuscake-paused                   | Pause the service                                |
| monitor.stakater.com/statuscake-ping-url                 | Webhook for alerts                               |
| monitor.stakater.com/statuscake-follow-redirect          | Enable ingress redirects                         |
| monitor.stakater.com/statuscake-port                     | TCP Port                                         |
| monitor.stakater.com/statuscake-trigger-rate             | Minutes to wait before sending an alert          |
| monitor.stakater.com/statuscake-contact-group            | Contact Group to be alerted.                     |