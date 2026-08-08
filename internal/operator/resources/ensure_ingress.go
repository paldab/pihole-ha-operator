package resources

import (
	"github.com/paldab/pihole-ha-operator/internal/operator/builders"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func applyIngressMetaData(current, desired *networkingv1.Ingress) {
	current.Labels = desired.Labels
	current.Annotations = desired.Annotations
}

func EnsureIngress(rc *ResourceContext) error {
	ingressConfig := rc.Cluster.Spec.Ingress

	if !*ingressConfig.Enabled {
		return nil
	}

	currentIng := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        rc.Cluster.Name,
			Namespace:   rc.Cluster.Namespace,
			Annotations: rc.Cluster.Spec.Ingress.Annotations,
		},
	}

	desiredIng := builders.BuildIngress(rc.Cluster)
	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, currentIng, func() error {

		applyIngressMetaData(currentIng, desiredIng)

		currentIng.Spec = desiredIng.Spec

		return ctrl.SetControllerReference(rc.Cluster, currentIng, rc.Scheme)
	})

	if err != nil {
		return err
	}

	return nil
}
