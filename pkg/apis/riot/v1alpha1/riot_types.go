package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RiotSpec defines the desired state of Riot
type RiotSpec struct {
	Replicas   int    `json:"replicas"`
	Image      string `json:"image"`
	ServerName string `json:"serverName"`
	Config     string `json:"config"`
}

// RiotStatus defines the observed state of Riot
type RiotStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Riot is the Schema for the riots API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=riots,scope=Namespaced
type Riot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RiotSpec   `json:"spec,omitempty"`
	Status RiotStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RiotList contains a list of Riot
type RiotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Riot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Riot{}, &RiotList{})
}
