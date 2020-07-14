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
	LocalObjectReference string `json:"name"`
}

// RouteURLSource selects a Route to populate the URL with
type RouteURLSource struct {
	LocalObjectReference string `json:"name"`
}

// LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.
type LocalObjectReference struct {
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
