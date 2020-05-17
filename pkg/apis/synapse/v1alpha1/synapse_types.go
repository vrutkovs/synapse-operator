package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SynapseConfig contains homeserver configuration
type SynapseConfig struct {
	Homeserver string `json:"homeserver"`
	Logging    string `json:"logging"`
}

// SynapseSecrets contains all secrets for synapse
type SynapseSecrets struct {
	Cert       string `json:"cert"`
	Key        string `json:"key"`
	SigningKey string `json:"signingKey"`
}

// SynapseSpec defines the desired state of Synapse
type SynapseSpec struct {
	Image      string         `json:"image"`
	ServerName string         `json:"serverName"`
	Config     SynapseConfig  `json:"configuration"`
	Secrets    SynapseSecrets `json:"secrets"`
}

// SynapseStatus defines the observed state of Synapse
type SynapseStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Synapse is the Schema for the synapses API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=synapses,scope=Namespaced
type Synapse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SynapseSpec   `json:"spec,omitempty"`
	Status SynapseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SynapseList contains a list of Synapse
type SynapseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Synapse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Synapse{}, &SynapseList{})
}
