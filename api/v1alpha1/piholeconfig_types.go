/*
Copyright 2026.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PiHoleConfigSpec defines the desired state of PiHoleConfig
type PiHoleConfigSpec struct {
	// ClusterRef references a PiHoleCluster in the same namespace.
	ClusterRef ClusterRef `json:"clusterRef"`

	Adlists AdList `json:"adlists,omitempty"`

	Denylist List `json:"denylist,omitempty"`

	Allowlist List `json:"allowlist,omitempty"`

	Regexlist RegexList `json:"regexlist,omitempty"`

	CNAMEs CNAMERecords `json:"cnames,omitempty"`

	Hosts HostRecords `json:"hosts,omitempty"`

	CustomOptions CustomOptions `json:"customOptions,omitempty"`
}

// PiHoleConfigStatus defines the observed state of PiHoleConfig.
type PiHoleConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PiHoleConfig resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	ActiveAdlists   int32 `json:"activeAdlists"`
	AdditionalHosts int32 `json:"additionalHosts"`
	CNAMES          int32 `json:"cnames"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PiHoleConfig is the Schema for the piholeconfigs API
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterRef.name"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].reason"
type PiHoleConfig struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PiHoleConfig
	// +required
	Spec PiHoleConfigSpec `json:"spec"`

	// status defines the observed state of PiHoleConfig
	// +optional
	Status PiHoleConfigStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PiHoleConfigList contains a list of PiHoleConfig
type PiHoleConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PiHoleConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PiHoleConfig{}, &PiHoleConfigList{})
		return nil
	})
}
