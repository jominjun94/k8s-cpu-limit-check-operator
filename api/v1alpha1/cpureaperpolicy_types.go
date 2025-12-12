package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CpuReaperPolicySpec defines the desired state
type CpuReaperPolicySpec struct {
	PodSelector          *metav1.LabelSelector `json:"podSelector,omitempty"`
	ThresholdPercent     int                   `json:"thresholdPercent,omitempty"`
	ForSeconds           int                   `json:"forSeconds,omitempty"`
	CheckIntervalSeconds int                   `json:"checkIntervalSeconds,omitempty"`
	RequireLimits        bool                  `json:"requireLimits,omitempty"`
}

// CpuReaperPolicyStatus defines the observed state
type CpuReaperPolicyStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type CpuReaperPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CpuReaperPolicySpec   `json:"spec,omitempty"`
	Status CpuReaperPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type CpuReaperPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CpuReaperPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CpuReaperPolicy{}, &CpuReaperPolicyList{})
}
