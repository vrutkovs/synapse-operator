package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SynapseWorkerSpec defines the desired state of SynapseWorker
type SynapseWorkerSpec struct {
	Replicas  int                     `json:"replicas"`
	Synapse   string                  `json:"synapse"`
	Worker    string                  `json:"worker"`
	Protocol  string                  `json:"protocol"`
	Port      int                     `json:"port"`
	Resources []SynapseWorkerResource `json:"resources"`
}

// SynapseWorkerResource defines synapse worker
type SynapseWorkerResource struct {
	Names []string `json:"names"`
}

// SynapseWorkerStatus defines the observed state of SynapseWorker
type SynapseWorkerStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SynapseWorker is the Schema for the synapseworkers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=synapseworkers,scope=Namespaced
type SynapseWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SynapseWorkerSpec   `json:"spec,omitempty"`
	Status SynapseWorkerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SynapseWorkerList contains a list of SynapseWorker
type SynapseWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SynapseWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SynapseWorker{}, &SynapseWorkerList{})
}
