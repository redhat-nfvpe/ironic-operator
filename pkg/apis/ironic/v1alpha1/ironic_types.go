package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IronicSpec defines the desired state of Ironic
type IronicSpec struct {
    Size int32 `json:"size"`
}

// IronicStatus defines the observed state of Ironic
type IronicStatus struct {
    Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Ironic is the Schema for the ironics API
// +k8s:openapi-gen=true
type Ironic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IronicSpec   `json:"spec,omitempty"`
	Status IronicStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IronicList contains a list of Ironic
type IronicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ironic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ironic{}, &IronicList{})
}
