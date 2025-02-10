package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CascadeRunSpec struct {

	Ob string `json:"ob"`
	Src string `json:"src"`
	PID string `json:"pid"`
	ScenarioName string `json:"scenarioname"`
	Modules []string `json:"modules"`
}

type CascadeRunStatus struct {
	Result []string `json:"result"`
	Info string `json:"info"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Job Result",type="string",JSONPath=".status.result",description="Jobs result"
// +kubebuilder:printcolumn:name="Info",type="string",JSONPath=".status.info",description="Information"
// CascadeAutoOperator is the Schema for the cascadeautooperators API
type CascadeRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CascadeRunSpec   `json:"spec,omitempty"`
	Status CascadeRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CascadeAutoOperatorList contains a list of CascadeAutoOperator
type CascadeRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CascadeRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CascadeRun{}, &CascadeRunList{})
}