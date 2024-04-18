/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CascadeModule struct {
	// Configuration parameter for Cascade Module
	// +patchMergeKey=name
	// +patchStrategy=merge
	ModuleName    string            `json:"modulename"`
	Configuration map[string]string `json:"configuration"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies the duration in seconds relative to the startTime that the job may be active
	// before the system tries to terminate it; value must be positive integer
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty" protobuf:"varint,3,opt,name=activeDeadlineSeconds"`

	// Specifies the number of retries before marking this job failed.
	// Defaults to 6
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty" protobuf:"varint,7,opt,name=backoffLimit"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Complete or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, it is eligible to be
	// automatically deleted. When the Job is being deleted, its lifecycle
	// guarantees (e.g. finalizers) will be honored. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// This field is alpha-level and is only honored by servers that enable the
	// TTLAfterFinished feature.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty" protobuf:"varint,8,opt,name=ttlSecondsAfterFinished"`

	// Describes the pod that will be created when executing a job.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	Template corev1.PodTemplateSpec `json:"template" protobuf:"bytes,6,opt,name=template"`
}

type CascadeScenario struct {
	// CascadeModules list with configuration parameters
	// +patchMergeKey=name
	// +patchStrategy=merge
	CascadeModules []CascadeModule `json:"cascademodules"`
}

// CascadeAutoOperatorSpec defines the desired state of CascadeAutoOperator
type CascadeAutoOperatorSpec struct {

	//Job configuration for Scenario
	ScenarioConfig CascadeScenario `json:"scenarioconfig"`

	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this deployment.
	// It must match the pod template's labels.
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	// Template describes the pods that will be created.
	Template corev1.PodTemplateSpec `json:"template" protobuf:"bytes,3,opt,name=template"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	Strategy apps.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,5,opt,name=minReadySeconds"`

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	// Indicates that the deployment is paused.
	// +optional
	Paused bool `json:"paused,omitempty" protobuf:"varint,7,opt,name=paused"`

	// The maximum time in seconds for a deployment to make progress before it
	// is considered to be failed. The deployment controller will continue to
	// process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Note that progress will
	// not be estimated during the time a deployment is paused. Defaults to 600s.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`
}

// CascadeAutoOperatorStatus defines the observed state of CascadeAutoOperator
type CascadeAutoOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Active    int32  `json:"active"`
	Succeeded int32  `json:"succeeded"`
	Failed    int32  `json:"failed"`
	Result    string `json:"result"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Active Jobs",type="string",JSONPath=".status.active",description="The active status of this Scenario"
// +kubebuilder:printcolumn:name="Succeeded Jobs",type="string",JSONPath=".status.succeeded",description="The succeeded status of this Scenario"
// +kubebuilder:printcolumn:name="Failed Jobs",type="string",JSONPath=".status.failed",description="The failed status of this Scenario"
// +kubebuilder:printcolumn:name="Last Scenario Result",type="string",JSONPath=".status.result",description="The result of last scenario run"
// CascadeAutoOperator is the Schema for the cascadeautooperators API
type CascadeAutoOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CascadeAutoOperatorSpec   `json:"spec,omitempty"`
	Status CascadeAutoOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CascadeAutoOperatorList contains a list of CascadeAutoOperator
type CascadeAutoOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CascadeAutoOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CascadeAutoOperator{}, &CascadeAutoOperatorList{})
}
