package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type PiholePodStorage struct {
	Size         *resource.Quantity `json:"size,omitempty"`
	StorageClass *string            `json:"storageClass,omitempty"`
}

type Probes struct {
	Startup   *corev1.Probe `json:"startup,omitempty"`
	Readiness *corev1.Probe `json:"readiness,omitempty"`
	Liveness  *corev1.Probe `json:"liveness,omitempty"`
}

type PiHolePodConfig struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	Env []corev1.EnvVar `json:"env,omitempty"`

	Affinity     *corev1.Affinity    `json:"affinity,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`

	Resources          corev1.ResourceRequirements `json:"resources,omitempty"`
	Probes             *Probes                     `json:"probes,omitempty"`
	SecurityContext    *corev1.SecurityContext     `json:"securityContext,omitempty"`
	PodSecurityContext *corev1.PodSecurityContext  `json:"podSecurityContext,omitempty"`
}

// ServiceConfig defines configuration for a generated Kubernetes Service.
// +kubebuilder:validation:XValidation:rule="!has(self.nodePort) || (has(self.type) && self.type in ['NodePort', 'LoadBalancer'])",message="nodePort may only be set when type is NodePort or LoadBalancer"
// +kubebuilder:validation:XValidation:rule="!has(self.loadBalancerIP) || (has(self.type) && self.type == 'LoadBalancer')",message="loadBalancerIP may only be set when type is LoadBalancer"
type ServiceConfig struct {
	// +optional
	Enabled *bool `json:"enabled"`

	// +optional
	Annotations map[string]string `json:"annotations"`

	// +optional
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type *corev1.ServiceType `json:"type,omitempty"`

	// Only valid when type is NodePort or LoadBalancer.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	NodePort *int32 `json:"nodePort"`

	// +optional
	// +kubebuilder:validation:Format=ipv4
	LoadBalancerIP *string `json:"loadBalancerIP"`
}

type PiHoleServiceSpec struct {
	Web  *ServiceConfig `json:"web,omitempty"`
	DNS  *ServiceConfig `json:"dns,omitempty"`
	DHCP *ServiceConfig `json:"dhcp,omitempty"`
}

type PiHoleIngressSpec struct {
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled"`

	Annotations map[string]string `json:"annotations,omitempty"`

	// Hostname used to access the Pi-hole web interface.
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host"`

	ClassName *string `json:"className"`

	TLS *PiHoleIngressTLS `json:"tls,omitempty"`
}

type PiHoleIngressTLS struct {
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName"`
}

type StatisticsSyncConfig struct {
	Mode     StatsMode            `json:"mode"`
	External *ExternalStatsConfig `json:"external"`
}

type StatsMode string

const (
	StatsModeLocal    StatsMode = "local"
	StatsModeExternal StatsMode = "external"
)

type ExternalStatsConfig struct {
	// +kubebuilder:default:=60
	ExportIntervalSeconds int `json:"exportIntervalSeconds"`
	BatchSize             int `json:"batchSize"`
}
