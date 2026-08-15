package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ExistingPasswordSecretRef struct {
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName"`

	// +kubebuilder:default:="password"
	PasswordKey *string `json:"passwordKey,omitempty"`
}

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
	Enabled *bool `json:"enabled,omitempty"`

	Annotations map[string]string `json:"annotations,omitempty"`

	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type *corev1.ServiceType `json:"type,omitempty"`

	// Only valid when type is NodePort or LoadBalancer.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	NodePort *int32 `json:"nodePort,omitempty"`

	// +kubebuilder:validation:Format=ipv4
	LoadBalancerIP *string `json:"loadBalancerIP,omitempty"`
}

type PiHoleServiceSpec struct {
	Web  *ServiceConfig `json:"web,omitempty"`
	DNS  *ServiceConfig `json:"dns,omitempty"`
	DHCP *ServiceConfig `json:"dhcp,omitempty"`
}

// PiHoleIngressSpec contains configuration for the ingress for the web interface
// +kubebuilder:validation:XValidation:rule="!has(self.enabled) || !self.enabled || has(self.host)",message="host is required when ingress is enabled"
type PiHoleIngressSpec struct {
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled"`

	Annotations map[string]string `json:"annotations,omitempty"`

	// Hostname used to access the Pi-hole web interface.
	// +kubebuilder:validation:MinLength=1
	Host *string `json:"host,omitempty"`

	ClassName *string `json:"className,omitempty"`

	TLS *PiHoleIngressTLS `json:"tls,omitempty"`
}

type StatisticsMode string

const (
	StatsModeLocal    StatisticsMode = "Local"
	StatsModeExternal StatisticsMode = "External"
)

type PiHoleIngressTLS struct {
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName"`
}

// StatisticsSpec contains the configuration of exporting statistics to an external DB or keep data local
// +kubebuilder:validation:XValidation:rule="self.mode != 'External' || has(self.external)",message="the external property is required when statistics.mode is External"
// +kubebuilder:validation:XValidation:rule="self.mode == 'External' || !has(self.external)",message="the external property may only be set when statistics.mode is External"
type StatisticsSpec struct {
	// +kubebuilder:validation:Enum=Local;External
	Mode StatisticsMode `json:"mode"`

	External *ExternalStatisticsConfig `json:"external,omitempty"`
}

type ExternalStatisticsConfig struct {
	// +kubebuilder:default:=60
	// +kubebuilder:validation:Minimum=60
	IntervalSeconds int32 `json:"intervalSeconds,omitempty"`

	// +kubebuilder:default:=100000
	// +kubebuilder:validation:Minimum=1
	BatchSize int32 `json:"batchSize,omitempty"`

	Database StatisticsDatabaseSpec `json:"database"`
}

type StatisticsDatabaseSpec struct {
	Host string `json:"host"`

	// +kubebuilder:default:=5432
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:default:="pihole_statistics"
	// +kubebuilder:validation:MinLength=1
	DBName string `json:"dbName,omitempty"`

	SecretRef StatsDatabaseSecretSelector `json:"secretRef"`
}

type StatsDatabaseSecretSelector struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +kubebuilder:default:="username"
	// +kubebuilder:validation:MinLength=1
	UsernameKey string `json:"usernameKey,omitempty"`

	// +kubebuilder:default:="password"
	// +kubebuilder:validation:MinLength=1
	PasswordKey string `json:"passwordKey,omitempty"`
}

type ExternalStatisticsStatus struct {
	IntervalSeconds int32 `json:"intervalSeconds"`
	BatchSize       int32 `json:"batchSize"`
}

type StatisticsStatus struct {
	Mode     StatisticsMode            `json:"mode"`
	External *ExternalStatisticsStatus `json:"external,omitempty"`
}
