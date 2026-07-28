package resources

import (
	"github.com/paldab/pihole-ha-operator/internal/builders"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func EnsureIngress(rc *ResourceContext) error {
	ingressConfig := rc.Cluster.Spec.Ingress

	if !*ingressConfig.Enabled {
		return nil
	}

	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        rc.Cluster.Name,
			Namespace:   rc.Cluster.Namespace,
			Annotations: rc.Cluster.Spec.Ingress.Annotations,
		},
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, ing, func() error {
		desired := builders.BuildIngress(rc.Cluster)

		ing.Labels = desired.Labels
		ing.Annotations = desired.Annotations
		ing.Spec = desired.Spec

		return ctrl.SetControllerReference(rc.Cluster, ing, rc.Scheme)
	})

	if err != nil {
		return err
	}

	return nil
}
