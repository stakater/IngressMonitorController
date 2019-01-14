# UptimeRobot Configuration
## Fetching alert contacts from UpTime Robot

In order to use Ingress Monitor controller, you need to have alert contacts added to your account. Once you add them via Dashboard, you will need their ID's. Fetching ID's is not something you can do via UpTime Robot's Dashboard. You will have to use their REST API to fetch alert contacts. To do that, run the following curl command on your terminal with your api key:

```bash
curl -d "api_key=your_api_key" -X POST https://api.uptimerobot.com/v2/getAlertContacts
```

You will get a response similar to what is shown below

```json
[
  {
    "stat": "ok",
    "offset": 0,
    "limit": 50,
    "total": 1,
    "alert_contacts": [
      {
        "id": "123456",
        "friendly_name": "hello",
        "type": 2,
        "status": 2,
        "value": "test@test.com"
      }
    ]
  }
]
```

Copy values of `id` field of your alert contacts which you want to use for Ingress Monitor Controller and append `_0_0` to them and seperate them by `-`. You will now have a string similar to `12345_0_0-23564_0_0`. This is basically the value you will need to specify in Ingress Monitor Controller's ConfigMap as `alertContacts`.

## Annotations

Additional uptime robot configurations can be added through a set of annotations to each ingress object, the current supported annotations are:

|                        Annotation             |                    Description                               |
|:---------------------------------------------:|:------------------------------------------------------------:|
| uptimerobot.monitor.stakater.com/interval     | The uptimerobot check interval in seconds                    |
| uptimerobot.monitor.stakater.com/status-pages | The uptimerobot public status page ID to add this monitor to |

### Fetching public status page ids from UpTime Robot

In order to use public status pages with the Ingress Monitor Controller you will need to have create one via the user interface.

You can then use their REST API to fetch the public status page id. To do that, run the following curl command on your terminal with your api key:

```bash
curl -d "api_key=your_api_key" -X POST https://api.uptimerobot.com/v2/getPsps
```

You will get a response similar to what is shown below

```json
{
  "stat": "ok",
  "pagination":
  {
    "offset": 0,
    "limit": 50,
    "total": 1
  },
  "psps":
  [
    {
      "id": 12345,
      "friendly_name": "my-public-status-page",
      "monitors": 0,
      "sort": 1,
      "status": 1,
      "standard_url": "https://stats.uptimerobot.com/12345678",
      "custom_url": ""
    }
  ]
}
```

Copy values of `id` field of your public status page which you want to use for Ingress Monitor Controller into the relevant ingress annotation.
