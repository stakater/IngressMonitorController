# Testing

## Running Tests Locally

Tests require a Kubernetes instance to talk to with a `test` namespace created, and a config with a valid UptimeRobot `apiKey` and `alertContacts`. For example, on MacOS with Homebrew and Minikube, you could accomplish this like

```bash
# install dependencies
$ brew install glide
$ glide update

# while still in the root folder, configure test setup
$ export CONFIG_FILE_PATH=$(pwd)/configs/testConfigs/test-config.yaml
# update the apikey and alertContacts in this file and the config_test.go file (`correctTestAPIKey` and `correctTestAlertContacts` contstants)
$ minikube start
$ kubectl create namespace test

# run the following command in the root folder
$ make test
```

## Test config for monitors

When running monitor test cases, make sure to provide a config similar to the following, making sure that the order of providers is the same as below:

```yaml
providers:
  - name: UptimeRobot
    apiKey: <your-api-key>
    apiURL: https://api.uptimerobot.com/v2/
    alertContacts: <your-alert-contacts>
  - name: StatusCake
    apiKey: <your-api-key>
    apiURL: https://app.statuscake.com/API/
    username: <your-account-username>
    password: <your-account-password>
  - name: Pingdom
    apiKey: <your-api-key>
    apiURL: https://api.pingdom.com
    username: <your-account-username>
    password: <your-account-password>
    accountEmail: <multi-auth-account-email>
enableMonitorDeletion: true
monitorNameTemplate: "{{.IngressName}}-{{.Namespace}}"
```

For example if you want to run only test cases for `StatusCake`, the 1st block of provider should still be present since test cases are written in that way.
