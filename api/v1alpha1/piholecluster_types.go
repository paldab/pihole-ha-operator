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
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// +kubebuilder:validation:MinLength=1
	ExistingAdminPasswordSecret string `json:"existingAdminPasswordSecret"`

	// +kubebuilder:default:=3
	Replicas *int32 `json:"replicas,omitempty"`

	Storage *PiholePodStorage `json:"storage,omitempty"`

	// +kubebuilder:default:="UTC"
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
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PiHoleCluster resource.
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

	// +optional
	CurrentLeader *string `json:"currentPrimary,omitempty"`

	Phase string `json:"phase"`

	// +kubebuilder:default:=false
	FailoverInProgress *bool `json:"failoverInProgress"`

	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`
	ReadyReplicas   int32 `json:"readyReplicas,omitempty"`
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`
	CurrentReplicas int32 `json:"currentReplicas"`

	Statistics StatisticsStatus `json:"statistics"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PiHoleCluster is the Schema for the piholeclusters API
// +kubebuilder:printcolumn:name="Instances",type=number,JSONPath=`.status.currentReplicas`
// +kubebuilder:printcolumn:name="Ready",type=number,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Primary",type=string,JSONPath=`.status.currentPrimary`
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
