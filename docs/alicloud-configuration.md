# Gcloud Configuration

You can configure AliCloud Cloud Monitoring Uptime Checks as a Ingress Monitor by using below configuration:


| Key          | Description                                                                        |
| -------------|------------------------------------------------------------------------------------|
| name         | Name of the provider (e.g. AliCloud)                                               |
| apiKey       | Access Key ID (AKID): This is a unique identifier for your Alibaba Cloud account. It is used in conjunction with the Access Key Secret to sign requests.|
| apiToken | Access Key Secret (AKSK): This is a secret key associated with the Access Key ID. It is used to sign requests to ensure that they are sent by a legitimate user. |
| apiURL | `apiUrl` refers to: https://api.aliyun.com/product/Cms                             |

When you create an Alibaba Cloud account, you are provided with an Access Key ID (AKID) and an Access Key Secret (AKSK). These credentials are used to sign requests to Alibaba Cloud APIs, ensuring that the requests are securely authenticated.

AccessKey ID and AccessKey Secret are your security credentials to access API of Alibaba Cloud, have full access to your account. Keep the AccessKey confidential.

**Example Configuration:**

```yaml
providers:
  - name: AliCloud
    apiKey: <ACCESS KEY>
    apiToken: <SECRET KEY>
    apiURL: "metrics.cn-qingdao.aliyuncs.com"
```
