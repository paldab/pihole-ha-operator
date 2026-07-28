package resources

import (
	"github.com/paldab/pihole-ha-operator/internal/builders"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func EnsureStatefulSet(rc *ResourceContext) error {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Cluster.Name,
			Namespace: rc.Cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrPatch(rc.Ctx, rc.K8sClient, sts, func() error {
		isNewObject := sts.CreationTimestamp.IsZero()
		desired := builders.BuildPiholeStatefulSet(rc.Cluster)

		sts.Labels = desired.Labels
		sts.Annotations = desired.Annotations

		if isNewObject {
			sts.Spec = desired.Spec
		} else {
			// Only update mutable fields
			sts.Spec.Replicas = desired.Spec.Replicas
			sts.Spec.Template = desired.Spec.Template
			sts.Spec.UpdateStrategy = desired.Spec.UpdateStrategy
			sts.Spec.PersistentVolumeClaimRetentionPolicy = desired.Spec.PersistentVolumeClaimRetentionPolicy
		}

		return ctrl.SetControllerReference(rc.Cluster, sts, rc.K8sClient.Scheme())
	})

	return err
}
