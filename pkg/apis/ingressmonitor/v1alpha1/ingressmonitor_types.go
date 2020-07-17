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
