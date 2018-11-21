package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IronicConductorSpec defines the desired state of IronicConductor
type IronicConductorSpec struct {
    Size int32 `json:"size"`
}

// IronicConductorStatus defines the observed state of IronicConductor
type IronicConductorStatus struct {
    Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IronicConductor is the Schema for the ironicconductors API
// +k8s:openapi-gen=true
type IronicConductor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IronicConductorSpec   `json:"spec,omitempty"`
	Status IronicConductorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IronicConductorList contains a list of IronicConductor
type IronicConductorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IronicConductor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IronicConductor{}, &IronicConductorList{})
}
