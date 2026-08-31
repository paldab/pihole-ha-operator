package resources

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/builders"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceBuilderFunction func(*v1alpha1.PiHoleCluster) *corev1.Service

type DesiredServiceConfig struct {
	Annotations            map[string]string
	Config                 v1alpha1.ServiceConfig
	Ports                  []corev1.ServicePort
	LoadBalancerIP         *string
	NodePort               *int32
	SessionAffinityTimeout *int32
}

func buildServiceMap(cluster *v1alpha1.PiHoleCluster) map[string]DesiredServiceConfig {
	serviceConfig := cluster.Spec.Services

	return map[string]DesiredServiceConfig{
		"web": {
			Config:         *serviceConfig.Web,
			Annotations:    cluster.Spec.Services.Web.Annotations,
			NodePort:       cluster.Spec.Services.Web.NodePort,
			LoadBalancerIP: cluster.Spec.Services.Web.LoadBalancerIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.Protocol("TCP"),
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       "https",
					Protocol:   corev1.Protocol("TCP"),
					Port:       443,
					TargetPort: intstr.FromInt(443),
				},
			},
			// SessionAffinityTimeout: new(int32(10800)),
		},

		"dns": {
			Config:         *serviceConfig.DNS,
			Annotations:    cluster.Spec.Services.DNS.Annotations,
			NodePort:       cluster.Spec.Services.DNS.NodePort,
			LoadBalancerIP: cluster.Spec.Services.DNS.LoadBalancerIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "dns",
					Protocol:   corev1.Protocol("TCP"),
					Port:       53,
					TargetPort: intstr.FromInt(53),
				},
				{
					Name:       "dns-udp",
					Protocol:   corev1.Protocol("UDP"),
					Port:       53,
					TargetPort: intstr.FromInt(53),
				},
			},
		},

		"dhcp": {
			Config:         *serviceConfig.Web,
			Annotations:    cluster.Spec.Services.DHCP.Annotations,
			NodePort:       cluster.Spec.Services.DHCP.NodePort,
			LoadBalancerIP: cluster.Spec.Services.DHCP.LoadBalancerIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "client-dhcp",
					Port:       67,
					Protocol:   corev1.Protocol("UDP"),
					TargetPort: intstr.FromInt(67),
				},
			},
		},
	}
}

func applyServiceMetaData(current, desired *corev1.Service) {
	current.Labels = utils.MergeMap(current.Labels, desired.Labels)
	current.Annotations = utils.MergeMap(current.Annotations, desired.Annotations)
}

func mergeServicePorts(current, desired []corev1.ServicePort) []corev1.ServicePort {
	results := make([]corev1.ServicePort, len(desired))
	copy(results, desired)

	for idx := range desired {
		for _, existingPort := range current {
			if results[idx].Name == existingPort.Name &&
				results[idx].Protocol == existingPort.Protocol &&
				results[idx].Port == existingPort.Port &&
				results[idx].NodePort == 0 {
				results[idx].NodePort = existingPort.NodePort

				break
			}
		}
	}

	return results
}

func EnsureServices(rc *ResourceContext) error {
	desiredServiceMap := buildServiceMap(rc.Cluster)
	serviceConfig := rc.Cluster.Spec.Services
	primaryPodLabels := utils.MergeMap(defaults.PiholeOperatorLabels(rc.Cluster.Name), defaults.PrimaryPodLabels)

	if !*serviceConfig.Web.Enabled {
		delete(desiredServiceMap, "web")
	}

	if !*serviceConfig.DNS.Enabled {
		delete(desiredServiceMap, "dns")
	}

	if !*serviceConfig.DHCP.Enabled {
		delete(desiredServiceMap, "dhcp")
	}

	for svcName, desiredServiceSettings := range desiredServiceMap {
		currentSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%s", rc.Cluster.Name, svcName),
				Namespace:   rc.Cluster.Namespace,
				Annotations: desiredServiceSettings.Annotations,
			},
		}

		desiredSvc := builders.BuildService(rc.Cluster, desiredServiceSettings.Config, desiredServiceSettings.Ports, primaryPodLabels)
		_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, currentSvc, func() error {
			applyServiceMetaData(currentSvc, desiredSvc)

			currentSvc.Spec.Type = desiredSvc.Spec.Type
			currentSvc.Spec.Selector = desiredSvc.Spec.Selector
			currentSvc.Spec.Ports = mergeServicePorts(currentSvc.Spec.Ports, desiredServiceSettings.Ports)

			if desiredSvc.Spec.LoadBalancerIP != "" {
				currentSvc.Spec.LoadBalancerIP = desiredSvc.Spec.LoadBalancerIP
			}

			return ctrl.SetControllerReference(rc.Cluster, currentSvc, rc.K8sClient.Scheme())
		})

		if err != nil {
			return err
		}
	}

	return nil
}
