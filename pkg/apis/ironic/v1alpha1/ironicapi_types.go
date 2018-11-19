package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


// IronicApiSpec defines the desired state of IronicApi
type IronicApiSpec struct {
    Size int32 `json:"size"`
}

// IronicApiStatus defines the observed state of IronicApi
type IronicApiStatus struct {
    Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IronicApi is the Schema for the ironicapis API
// +k8s:openapi-gen=true
type IronicApi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IronicApiSpec   `json:"spec,omitempty"`
	Status IronicApiStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IronicApiList contains a list of IronicApi
type IronicApiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IronicApi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IronicApi{}, &IronicApiList{})
}
