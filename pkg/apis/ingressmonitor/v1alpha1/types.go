package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressMonitor defines the desired state of IngressMonitor
// +k8s:openapi-gen=true
type IngressMonitorSpec struct {
	// Name of service monitor
	Name     string `json:"name"`
	// URL to monitor
	URL      string `json:"url,omitempty"`
	// URL to monitor from either an ingress or route reference
	// +optional
	URLFrom *URLSource `json:"urlFrom,omitempty"`
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
	LocalObjectReference
}

// RouteURLSource selects a Route to populate the URL with
type RouteURLSource struct {
	LocalObjectReference
}

// LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.
type LocalObjectReference struct {
	Name string `json:"name"`
}


// IngressMonitorStatus is the observed state of a IngressMonitor resource
// +k8s:openapi-gen=true
type IngressMonitorStatus struct {
}

// IngressMonitor is a specification for a IngressMonitor resource
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type IngressMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressMonitorSpec   `json:"spec,omitempty"`
	Status IngressMonitorStatus `json:"status,omitempty"`
}


// IngressMonitorList is a list of IngressMonitor resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type IngressMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty""`

	Items []IngressMonitor `json:"items"`
}