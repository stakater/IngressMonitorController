# Gcloud Configuration

You can configure Google Cloud Monitoring Uptime Checks as a Ingress Monitor by using below configuration:

| Key          | Description                                                                         |
| -------------| ----------------------------------------------------------------------------------- |
| name         | Name of the provider (e.g. gcloud)                                                  |
| apiKey       | JSON Service Account Key                                                            |
| gcloudConfig | `gcloudConfig` is the configuration specific to gcloud Instance as mentioned below: |

## gcloud Configuration:

| Key       | Description                                                                                                    |
| --------- | ----------------------- |
| projectId | Google Cloud Project ID |

**Example Configuration:**

```yaml
providers:
  - name: gcloud
    apiKey: |
      {
        "type": "service_account",
        "project_id": "...",
        "private_key_id": "...",
        "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
        "client_email": "...iam.gserviceaccount.com",
        "client_id": "...",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/...iam.gserviceaccount.com"
      }
    gcloudConfig:
      projectId: project-name
enableMonitorDeletion: true
```

## Example: 

```yaml
apiVersion: endpointmonitor.stakater.com/v1alpha1
kind: EndpointMonitor
metadata:
  name: stakater
spec:
  forceHtpps: true
  url: https://stakater.com/
  gcloudConfig:
    projectId: stakater-project
```