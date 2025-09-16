# Appinsights Configuration

You can configure Application Insights as a Ingress Monitor by using below configuration:

| Key               | Description                                                                                   |
| ----------------- | --------------------------------------------------------------------------------------------- |
| name              | Name of the provider (e.g. AppInsights)                                                       |
| appInsightsConfig | `appInsightsConfig` is the configuration specific to Appinsights Instance as mentioned below: |

## Authentication

The AppInsights monitor uses [azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go) to authenticate and communicate to the Azure API.

> The [DefaultAzureCredential](https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/credential-chains#defaultazurecredential-overview) is an opinionated, preconfigured chain of credentials.
> It's designed to support many environments, along with the most common authentication flows and developer tools. In graphical form, the underlying chain looks like this:

It will automatically configure authentication in the following order, stopping when it finds a hit:

* Environment Variables
* Workload Identity
* Managed Identity
* Azure CLI
* Azure Developer CLI

Refer to the [DefaultAzureCredential documentation](https://learn.microsoft.com/en-us/azure/developer/go/sdk/authentication/credential-chains#defaultazurecredential-overview) for more details.


## Appinsights Configuration:

| Key                      | Description                                                                                                    |
|--------------------------|----------------------------------------------------------------------------------------------------------------|
| name                     | Name of the Appinsights Instance                                                                               |
| subscriptionId           | The Azure Subscription ID                                                                                      |
| resourceGroup            | Resource group of Appinsights                                                                                  |
| location                 | The location of the resource group.                                                                            |
| geoLocation              | Location ID for the webtest to run from. For example: `["us-tx-sn1-azr", "us-il-ch1-azr"]`. See [Azure documentation](https://learn.microsoft.com/en-us/previous-versions/azure/azure-monitor/app/monitor-web-app-availability#location-population-tags) for details on Location population tags. |
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
      name: "demo-appinsighs"
      subscriptionId: "12345678-1234-1234-1234-123456789012"
      resourceGroup: "demoRG"
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
        service_uri: "http://myalert-webhook.io"
enableMonitorDeletion: true
```

## Additional Configuration

Additional Appinsights configurations can be added in the `EndpointMonitor`, current supported configuration attributes are:

| Fields                                  | Description                                                                                                                                      |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| StatusCode  | Returned status code that is counted as a success. Possible values: [HTTP Status Codes](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes) |
| RetryEnable | If its `true`, falied test will be retry after a short interval. Possible values: `true, false`                                                  |
| Frequency   | Sets how often the test should run from each test location. Possible values: `300,600,900` seconds                                               |

## Example: 

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHttps: true
  url: https://stakater.com/
  appInsightsConfig:
    statusCode: 404
    retryEnable: true
    frequency: 900
```
