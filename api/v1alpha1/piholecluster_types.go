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

// PiHoleClusterSpec defines the desired state of PiHoleCluster
type PiHoleClusterSpec struct {
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	ExistingSecretRef ExistingPasswordSecretRef `json:"existingSecretRef"`

	// +kubebuilder:default:=3
	// +kubebuilder:validation:Minimum=1
	// Replicas minimum 1 because the operator is not build to handle 0 replicas
	Replicas *int32 `json:"replicas,omitempty"`

	Storage *PiholePodStorage `json:"storage,omitempty"`

	// +kubebuilder:default:="UTC"
	// +kubebuilder:validation:MinLength=1
	TimeZone *string `json:"timezone,omitempty"`

	// +kubebuilder:default:={}
	DNSUpstreams []string `json:"dnsUpstreams,omitempty"`

	Config *PiHolePodConfig `json:"config,omitempty"`

	Services *PiHoleServiceSpec `json:"services,omitempty"`

	Ingress *PiHoleIngressSpec `json:"ingress,omitempty"`

	Statistics *StatisticsSpec `json:"statistics,omitempty"`
}

// PiHoleClusterStatus defines the observed state of PiHoleCluster.
type PiHoleClusterStatus struct {
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	CurrentLeader *string `json:"currentPrimary,omitempty"`

	// todo: remove
	Phase string `json:"phase"`

	// +kubebuilder:default:=false
	// todo: remove
	FailoverInProgress *bool `json:"failoverInProgress"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`
	ReadyReplicas   int32 `json:"readyReplicas,omitempty"`
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	Statistics StatisticsStatus `json:"statistics"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PiHoleCluster is the Schema for the piholeclusters API
// +kubebuilder:printcolumn:name="Replicas",type=number,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Primary",type=string,JSONPath=`.status.currentPrimary`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Reason",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].reason"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type PiHoleCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PiHoleCluster
	// +required
	Spec PiHoleClusterSpec `json:"spec"`

	// status defines the observed state of PiHoleCluster
	// +optional
	Status PiHoleClusterStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PiHoleClusterList contains a list of PiHoleCluster
type PiHoleClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PiHoleCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PiHoleCluster{}, &PiHoleClusterList{})
		return nil
	})
}
