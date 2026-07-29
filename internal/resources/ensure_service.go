package resources

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/builders"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
					Name:     "http",
					Port:     80,
					Protocol: corev1.Protocol("TCP"),
				},
				{
					Name:     "https",
					Port:     443,
					Protocol: corev1.Protocol("TCP"),
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
					Name:     "dns",
					Port:     53,
					Protocol: corev1.Protocol("TCP"),
				},
				{
					Name:     "dns-udp",
					Port:     53,
					Protocol: corev1.Protocol("UDP"),
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
					Name:     "client-dhcp",
					Port:     67,
					Protocol: corev1.Protocol("UDP"),
				},
			},
		},
	}
}

func EnsureServices(rc *ResourceContext) error {
	desiredServiceMap := buildServiceMap(rc.Cluster)
	serviceConfig := rc.Cluster.Spec.Services
	primaryPodLabels := utils.MergeMap(defaults.PiholeOperatorLabels(rc.Cluster), defaults.PrimaryPodLabels)

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
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%s", rc.Cluster.Name, svcName),
				Namespace:   rc.Cluster.Namespace,
				Annotations: desiredServiceSettings.Annotations,
			},
		}

		_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, svc, func() error {
			desired := builders.BuildService(rc.Cluster, desiredServiceSettings.Config, desiredServiceSettings.Ports, primaryPodLabels)

			svc.Labels = desired.Labels
			svc.Annotations = desired.Annotations
			svc.Spec = desired.Spec

			return ctrl.SetControllerReference(rc.Cluster, svc, rc.K8sClient.Scheme())
		})

		if err != nil {
			return err
		}
	}

	return nil
}
