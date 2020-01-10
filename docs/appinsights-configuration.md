# Appinsights Configuration

You can configure Application Insights as a Ingress Monitor by using below configuration:

| Key               | Description                                                                                   |
| ----------------- | --------------------------------------------------------------------------------------------- |
| name              | Name of the provider (e.g. AppInsights)                                                       |
| appInsightsConfig | `appInsightsConfig` is the configuration specific to Appinsights Instance as mentioned below: |

## Appinsights Configuration:

| Key                      | Description                                                                                                    |
| ------------------------ | -------------------------------------------------------------------------------------------------------------- |
| name                     | Name of the Appinsights Instance                                                                               |
| resourceGroup            | Resource group of Appinsights                                                                                  |
| location                 | The location of the resource group.                                                                            |
| geoLocation              | Location ID for the webtest to run from. For example: `["us-tx-sn1-azr", "us-il-ch1-azr"]`                     |
| emailAction (Optional)   | Email Action is optional, This will enable monitoring alerts for ping test failure.                            |
| webhookAction (Optional) | Webhook Action is also optional, You can use webhooks to route an Azure alert notification for custom actions. |

**Email Action:**

- **send_to_service_owners**: send email to all service owners, Possible values: `true, false`
- **custom_emails**: list of email ids, For example: `["abc@microsoft.com", "xyz@microsoft.com"]`

**Webhook Action:**

- service_uri: Webhook url, For example: `http://webhook-test.io`

**Example Configuration:**

```yaml
providers:
  - name: AppInsights
    appInsightsConfig:
      name: demo-appinsights
      resourceGroup: demoRG
      location: "westeurope"
      geoLocation:
        [
          "us-tx-sn1-azr",
          "emea-nl-ams-azr",
          "us-fl-mia-edge",
          "latam-br-gru-edge",
        ]
      emailAction:
        send_to_service_owners: false
        custom_emails: ["mail@cizer.dev"]
      webhookAction:
        service_uri: http://myalert-webhook.io
enableMonitorDeletion: true
```

## Additional Configuration

Additional Appinsights configurations can be added to each ingress object using annotations, current supported annotations are:

| Annotations                                  | Description                                                                                                                                      |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| appinsights.monitor.stakater.com/statuscode  | Returned status code that is counted as a success. Possible values: [HTTP Status Codes](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes) |
| appinsights.monitor.stakater.com/retryenable | If its `true`, falied test will be retry after a short interval. Possible values: `true, false`                                                  |
| appinsights.monitor.stakater.com/frequency   | Sets how often the test should run from each test location. Possible values: `300,600,900` seconds                                               |
