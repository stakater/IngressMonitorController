/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EndpointMonitorSpec defines the desired state of EndpointMonitor
type EndpointMonitorSpec struct {
	// URL to monitor
	URL string `json:"url,omitempty"`

	// Force monitor endpoint to use HTTPS
	// +optional
	ForceHTTPS bool `json:"forceHttps,omitempty"`

	// +optional
	HealthEndpoint string `json:"healthEndpoint,omitempty"`

	// Comma separated list of providers
	// +optional
	Providers string `json:"providers"`

	// URL to monitor from either an ingress or route reference
	// +optional
	URLFrom *URLSource `json:"urlFrom,omitempty"`

	// Configuration for UptimeRobot Monitor Provider
	// +optional
	UptimeRobotConfig *UptimeRobotConfig `json:"uptimeRobotConfig,omitempty"`

	// Configuration for Uptime Monitor Provider
	// +optional
	UptimeConfig *UptimeConfig `json:"uptimeConfig,omitempty"`

	// Configuration for Updown Monitor Provider
	// +optional
	UpdownConfig *UpdownConfig `json:"updownConfig,omitempty"`

	// Configuration for StatusCake Monitor Provider
	// +optional
	StatusCakeConfig *StatusCakeConfig `json:"statusCakeConfig,omitempty"`

	// Configuration for Pingdom Monitor Provider
	// +optional
	PingdomConfig *PingdomConfig `json:"pingdomConfig,omitempty"`

	// Configuration for Pingdom Transaction Monitor Provider
	// +optional
	PingdomTransactionConfig *PingdomTransactionConfig `json:"pingdomTransactionConfig,omitempty"`

	// Configuration for AppInsights Monitor Provider
	// +optional
	AppInsightsConfig *AppInsightsConfig `json:"appInsightsConfig,omitempty"`

	// Configuration for Google Cloud Monitor Provider
	// +optional
	GCloudConfig *GCloudConfig `json:"gcloudConfig,omitempty"`

	// Configuration for Grafana Cloud Monitor Provider
	// +optional
	GrafanaConfig *GrafanaConfig `json:"grafanaConfig,omitempty"`
}

// UptimeRobotConfig defines the configuration for UptimeRobot Monitor Provider
type UptimeRobotConfig struct {
	// The uptimerobot alertContacts to be associated with this monitor
	// +optional
	AlertContacts string `json:"alertContacts,omitempty"`

	// The uptimerobot check interval in seconds
	// +kubebuilder:validation:Minimum=60
	// +optional
	Interval int `json:"interval,omitempty"`

	// Specify maintenanceWindows i.e. once or recurring “do-not-monitor periods”
	// +optional
	MaintenanceWindows string `json:"maintenanceWindows,omitempty"`

	// The uptimerobot monitor type (http or keyword)
	// +kubebuilder:validation:Enum=http;keyword
	// +optional
	MonitorType string `json:"monitorType,omitempty"`

	// Alert if value exist (yes) or doesn't exist (no) (Only if monitor-type is keyword)
	// +kubebuilder:validation:Enum=yes;no
	// +optional
	KeywordExists string `json:"keywordExists,omitempty"`

	// keyword to check on URL (e.g.'search' or '404') (Only if monitor-type is keyword)
	// +optional
	KeywordValue string `json:"keywordValue,omitempty"`

	// The uptimerobot public status page ID to add this monitor to
	// +optional
	StatusPages string `json:"statusPages,omitempty"`

	// Defines which http status codes are treated as up or down
	// For ex: 200:0_401:1_503:1 (to accept 200 as down and 401 and 503 as up)
	CustomHTTPStatuses string `json:"customHTTPStatuses,omitempty"`
}

// UptimeConfig defines the configuration for Uptime Monitor Provider
type UptimeConfig struct {
	// The uptime check interval in seconds
	// +optional
	Interval int `json:"interval,omitempty"`

	// The uptime check type that can be HTTP/DNS/ICMP etc.
	// +optional
	CheckType string `json:"checkType,omitempty"`

	// Add one or more contact groups separated by `,`
	// +optional
	Contacts string `json:"contacts,omitempty"`

	// Add different locations for the check
	// +optional
	Locations string `json:"locations,omitempty"`

	// Add one or more tags for the check separated by `,`
	// +optional
	Tags string `json:"tags,omitempty"`
}

// UpdownConfig defines the configuration for Updown Monitor Provider
type UpdownConfig struct {
	// Enable or disable checks
	// +optional
	Enable bool `json:"enable,omitempty"`

	// The pingdom check interval in seconds
	// +optional
	Period int `json:"period,omitempty"`

	// Make status page public or not
	// +optional
	PublishPage bool `json:"publishPage,omitempty"`

	// Additional request headers for API calls
	// +optional
	RequestHeaders string `json:"requestHeaders,omitempty"`
}

// StatusCakeConfig defines the configuration for StatusCake Monitor Provider
type StatusCakeConfig struct {
	// Basic Auth User
	// +optional
	BasicAuthUser string `json:"basicAuthUser,omitempty"`

	// Set Check Rate for the monitor
	// +optional
	CheckRate int `json:"checkRate,omitempty"`

	// Set Test type - HTTP, TCP, PING
	// +optional
	TestType string `json:"testType,omitempty"`

	// Pause the service
	// +optional
	Paused bool `json:"paused,omitempty"`

	// Webhook for alerts
	// +optional
	PingURL string `json:"pingUrl,omitempty"`

	// Enable ingress redirects
	// +optional
	FollowRedirect bool `json:"followRedirect,omitempty"`

	// TCP Port
	// +optional
	Port int `json:"port,omitempty"`

	// Minutes to wait before sending an alert
	// +optional
	TriggerRate int `json:"triggerRate,omitempty"`

	// Contact Group to be alerted.
	// +optional
	ContactGroup string `json:"contactGroup,omitempty"`

	// Comma separated list of tags
	// +optional
	TestTags string `json:"testTags,omitempty"`

	// Comma separated list of Node Location IDs
	// +optional
	Regions string `json:"regions,omitempty"`

	// Comma separated list of HTTP codes to trigger error on
	// +optional
	StatusCodes string `json:"statusCodes,omitempty"`

	// Confirmation value ranges from (0,10)
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:validation:Minimum=0
	// +optional
	Confirmation int `json:"confirmation,omitempty"`

	// Enable SSL Alert
	// +optional
	EnableSSLAlert bool `json:"enableSslAlert,omitempty"`

	// Enable Real Browser
	// +optional
	RealBrowser bool `json:"realBrowser,omitempty"`

	// String to look for within the response. Considered down if not found
	// +optional
	FindString string `json:"findString,omitempty"`

	// RawPostData can be used to send parameters within the URL. Changes the request from a GET to a POST
	// +optional
	RawPostData string `json:"rawPostData,omitempty"`
}

// PingdomConfig defines the configuration for Pingdom Monitor Provider
type PingdomConfig struct {
	// The pingdom check interval in minutes
	// +optional
	Resolution int `json:"resolution,omitempty"`

	// How many failed check attempts before notifying
	// +optional
	SendNotificationWhenDown int `json:"sendNotificationWhenDown,omitempty"`

	// Set to "true" to pause checks
	// +optional
	Paused bool `json:"paused,omitempty"`

	// Set to "false" to disable recovery notifications
	// +optional
	NotifyWhenBackUp bool `json:"notifyWhenBackUp,omitempty"`

	// Custom request headers
	// +optional
	RequestHeaders string `json:"requestHeaders,omitempty"`

	// Custom request headers that should be read from an environment variable as it possibly contains sensitive data.
	// An example would be an API token.
	// +optional
	RequestHeadersEnvVar string `json:"requestHeadersEnvVar,omitempty"`

	// Required for basic-authentication
	// +optional
	BasicAuthUser string `json:"basicAuthUser,omitempty"`

	// Set to text string that has to be present in the HTML code of the page
	// +optional
	ShouldContain string `json:"shouldContain,omitempty"`

	// Comma separated set of tags to apply to check (e.g. "testing,aws")
	// +optional
	Tags string `json:"tags,omitempty"`

	// `-` separated set list of integrations ids (e.g. "91166-12168")
	// +optional
	AlertIntegrations string `json:"alertIntegrations,omitempty"`

	// `-` separated contact id's (e.g. "1234567_8_9-9876543_2_1")
	// +optional
	AlertContacts string `json:"alertContacts,omitempty"`

	// `-` separated team id's (e.g. "1234567_8_9-9876543_2_1")
	// +optional
	TeamAlertContacts string `json:"teamAlertContacts,omitempty"`

	// Monitor SSL/TLS certificate
	// Monitor the validity of your SSL/TLS certificate. With this enabled Uptime checks will be considered DOWN when
	// the certificate becomes invalid or expires.
	// SSL/TLS certificate monitoring is available for HTTP checks.
	// +optional
	VerifyCertificate bool `json:"verifyCertificate,omitempty"`

	// Consider down prior to certificate expiring
	// Select the number of days prior to your certificate expiry date that you want to consider the check down.
	// At this day your check will be considered down and if applicable a down alert will be sent.
	// +optional
	SSLDownDaysBefore int `json:"sslDownDaysBefore,omitempty"`

	// Data that should be posted to the web page, for example submission data for a sign-up or login form.
	// The data needs to be formatted in the same way as a web browser would send it to the web server.
	// Because post data contains sensitive secret this field is only a reference to an environment variable.
	// +optional
	PostDataEnvVar string `json:"postDataEnvVar,omitempty"`
}

// PingdomTransactionConfig defines the configuration for Pingdom Transaction Monitor Provider
type PingdomTransactionConfig struct {

	// Check status: active or inactive
	// +optional
	Paused bool `json:"paused,omitempty"`

	// Custom message that is part of the email and webhook alerts
	// +optional
	CustomMessage string `json:"custom_message,omitempty"`

	// TMS test intervals in minutes. Allowed intervals: 5,10,20,60,720,1440. The interval you're allowed to set may vary depending on your current plan.
	// +optional
	// +kubebuilder:validation:Enum=5;10;20;60;720;1440
	Interval int `json:"interval,omitempty"`

	// Name of the region where the check is executed. Supported regions: us-east, us-west, eu, au
	// +optional
	// +kubebuilder:validation:Enum=us-east;us-west;eu;au
	Region string `json:"region,omitempty"`

	// Send notification when down X times
	SendNotificationWhenDown int64 `json:"send_notification_when_down,omitempty"`

	// Check importance- how important are the alerts when the check fails. Allowed values: low, high
	// +optional
	// +kubebuilder:validation:Enum=low;high
	SeverityLevel string `json:"severity_level,omitempty"`

	// steps to be executed as part of the check
	// +required
	Steps []PingdomStep `json:"steps"`

	// List of tags for a check. The tag name may contain the characters 'A-Z', 'a-z', '0-9', '_' and '-'. The maximum length of a tag is 64 characters.
	Tags []string `json:"tags,omitempty"`

	// `-` separated set list of integrations ids (e.g. "91166-12168")
	// +optional
	AlertIntegrations string `json:"alertIntegrations,omitempty"`

	// `-` separated contact id's (e.g. "1234567_8_9-9876543_2_1")
	// +optional
	AlertContacts string `json:"alertContacts,omitempty"`

	// `-` separated team id's (e.g. "1234567_8_9-9876543_2_1")
	// +optional
	TeamAlertContacts string `json:"teamAlertContacts,omitempty"`
}

// PingdomStep respresents a step of the script to run a transcaction check
type PingdomStep struct {
	// contains the html element with assigned value
	// the key element is always lowercase for example {"url": "https://www.pingdom.com"}
	// see available values at https://pkg.go.dev/github.com/karlderkaefer/pingdom-golang-client@latest/pkg/pingdom/client/tmschecks#StepArg
	// +required
	Args map[string]string `json:"args"`
	// contains the function that is executed as part of the step
	// commands: go_to, click, fill, check, uncheck, sleep, select_radio, basic_auth, submit, wait_for_element, wait_for_contains
	// validations: url, exists, not_exists, contains, not_contains, field_contains, field_not_contains, is_checked, is_not_checked, radio_selected, dropdown_selected, dropdown_not_selected
	// see updated list https://docs.pingdom.com/api/#section/TMS-Steps-Vocabulary/Script-transaction-checks
	// +required
	Function string `json:"function"`
}

// AppInsightsConfig defines the configuration for AppInsights Monitor Provider
type AppInsightsConfig struct {
	// Returned status code that is counted as a success
	// +optional
	StatusCode int `json:"statusCode,omitempty"`

	// If its `true`, falied test will be retry after a short interval. Possible values: `true, false`
	// +optional
	RetryEnable bool `json:"retryEnable,omitempty"`

	// Sets how often the test should run from each test location. Possible values: `300,600,900` seconds
	// +optional
	Frequency int `json:"frequency,omitempty"`
}

// GCloudConfiguration defines the configuration for Google Cloud Monitor Provider
type GCloudConfig struct {
	// Google Cloud Project ID
	// +optional
	ProjectId string `json:"projectId,omitempty"`
}

// GrafnaConfiguration defines the configuration for Grafana Cloud Monitor Provider
type GrafanaConfig struct {
	TenantId int64 `json:"tenantId,omitempty"`

	// The frequency value specifies how often the check runs in milliseconds
	Frequency int64 `json:"frequency,omitempty"`

	// Probes are the monitoring agents responsible for simulating user interactions with your web applications
	// or services. These agents periodically send requests to predefined URLs and record the responses,
	// checking for expected outcomes and measuring performance.
	Probes []string `json:"probes,omitempty"`

	// The alertSensitivity value defaults to none if there are no alerts or can be set to low, medium,
	// or high to correspond to the check alert levels.
	// +kubebuilder:validation:Enum=none;low;medium;high
	// +kubebuilder:default=none
	AlertSensitivity string `json:"alertSensitivity,omitempty"`
}

// URLSource represents the set of resources to fetch the URL from
type URLSource struct {
	// +optional
	IngressRef *IngressURLSource `json:"ingressRef,omitempty"`
	// +optional
	RouteRef *RouteURLSource `json:"routeRef,omitempty"`
}

// IngressURLSource selects an Ingress to populate the URL with
type IngressURLSource struct {
	Name string `json:"name"`
}

// RouteURLSource selects a Route to populate the URL with
type RouteURLSource struct {
	Name string `json:"name"`
}

// EndpointMonitorStatus defines the observed state of EndpointMonitor
type EndpointMonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EndpointMonitor is the Schema for the endpointmonitors API
type EndpointMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointMonitorSpec   `json:"spec,omitempty"`
	Status EndpointMonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EndpointMonitorList contains a list of EndpointMonitor
type EndpointMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EndpointMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EndpointMonitor{}, &EndpointMonitorList{})
}
