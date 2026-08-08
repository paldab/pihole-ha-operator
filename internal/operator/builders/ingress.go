package builders

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildIngress(cluster *v1alpha1.PiHoleCluster) *networkingv1.Ingress {
	webSvcName := fmt.Sprintf("%s-%s", cluster.Name, "web")
	ingressConfig := cluster.Spec.Ingress
	ingressWebRule := buildIngressWebRule(webSvcName)

	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: cluster.Spec.Ingress.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: cluster.Spec.Ingress.ClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host:             ingressConfig.Host,
					IngressRuleValue: ingressWebRule,
				},
			},
		},
	}

	if cluster.Spec.Ingress.TLS != nil {
		ing.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts: []string{
					ingressConfig.Host,
				},
				SecretName: ingressConfig.TLS.SecretName,
			},
		}
	}

	return ing
}

func buildIngressWebRule(ingressName string) networkingv1.IngressRuleValue {
	defaultPathPrefix := networkingv1.PathTypePrefix

	return networkingv1.IngressRuleValue{
		HTTP: &networkingv1.HTTPIngressRuleValue{
			Paths: []networkingv1.HTTPIngressPath{
				{
					Path:     "/",
					PathType: &defaultPathPrefix,
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: ingressName,
							Port: networkingv1.ServiceBackendPort{
								Name: "http",
							},
						},
					},
				},
			},
		},
	}
}
