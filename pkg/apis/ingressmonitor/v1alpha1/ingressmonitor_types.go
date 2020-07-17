package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IngressMonitorSpec defines the desired state of IngressMonitor
type IngressMonitorSpec struct {
	// URL to monitor
	URL string `json:"url,omitempty"`

	// Force monitor endpoint to use HTTPS
	// +optional
	ForceHTTPS bool `json:"forceHttps,omitempty"`

	// +optional
	HealthEndpoint string `json:"healthEndpoint,omitempty"`

	// Comma separated list of providers
	// +required
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
}

// UptimeRobotConfig defines the configuration for UptimeRobot Monitor Provider
type UptimeRobotConfig struct {
	// The uptimerobot alertContacts to be associated with this monitor
	// +optional
	AlertContacts string `json:"alertContacts,omitempty"`

	// The uptimerobot check interval in seconds
	// +optional
	Interval int `json:"interval,omitempty"`

	// Specify maintenanceWindows i.e. once or recurring “do-not-monitor periods”
	// +optional
	MaintenanceWindows string `json:"maintenanceWindows,omitempty"`

	// The uptimerobot monitor type (http or keyword)
	// +optional
	MonitorType string `json:"monitorType,omitempty"`

	// Alert if value exist (yes) or doesn't exist (no) (Only if monitor-type is keyword)
	// +optional
	KeywordExists string `json:"keywordExists,omitempty"`

	// keyword to check on URL (e.g.'search' or '404') (Only if monitor-type is keyword)
	// +optional
	KeywordValue string `json:"keywordExists,omitempty"`

	// The uptimerobot public status page ID to add this monitor to
	// +optional
	StatusPages string `json:"statusPages,omitempty"`
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
	NodeLocations string `json:"nodeLocations,omitempty"`

	// Comma seperated list of HTTP codes to trigger error on
	// +optional
	StatusCodes string `json:"statusCodes,omitempty"`

	// Confirmation value ranges from (0,10)
	// +optional
	Confirmation int `json:"confirmation,omitempty"`

	// Enable SSL Alert
	// +optional
	EnableSSLAlert bool `json:"enableSslAlert,omitempty"`

	// Enable Real Browser
	// +optional
	RealBrowser bool `json:"realBrowser,omitempty"`
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
	Paused bool `json:"resolution,paused"`

	// Set to "false" to disable recovery notifications
	// +optional
	NotifyWhenBackUp bool `json:"notifyWhenBackUp,omitempty"`

	// Custom pingdom request headers
	// +optional
	RequestHeaders string `json:"requestHeaders,omitempty"`

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

// IngressMonitorStatus defines the observed state of IngressMonitor
type IngressMonitorStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressMonitor is the Schema for the ingressmonitors API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=ingressmonitors,scope=Namespaced
type IngressMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressMonitorSpec   `json:"spec,omitempty"`
	Status IngressMonitorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressMonitorList contains a list of IngressMonitor
type IngressMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IngressMonitor{}, &IngressMonitorList{})
}
