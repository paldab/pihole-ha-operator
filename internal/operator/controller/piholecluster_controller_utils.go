package controller

import (
	"context"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/resources"
	"github.com/paldab/pihole-ha-operator/internal/operator/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func getActivePiholeConfig(rc *resources.ResourceContext) (*piholev1alpha1.PiHoleConfig, error) {
	var piholeConfigList piholev1alpha1.PiHoleConfigList
	if err := rc.K8sClient.List(
		rc.Ctx,
		&piholeConfigList,
		client.InNamespace(rc.Cluster.Namespace),
		client.MatchingFields{
			".spec.clusterRef.name": rc.Cluster.Name,
		},
	); err != nil {
		return nil, err
	}

	var activeConfig *piholev1alpha1.PiHoleConfig
	// If one condition of piholeconfig is ready SKIP create empty configs
	for i := range piholeConfigList.Items {
		config := &piholeConfigList.Items[i]

		if meta.IsStatusConditionTrue(
			config.Status.Conditions,
			status.ConditionConfigReady,
		) {
			activeConfig = config
			break
		}
	}

	return activeConfig, nil
}

func watchManagedServices(e event.UpdateEvent) bool {
	oldSvc, ok := e.ObjectOld.(*corev1.Service)
	if !ok {
		return true
	}

	newSvc, ok := e.ObjectNew.(*corev1.Service)
	if !ok {
		return true
	}

	return !equality.Semantic.DeepEqual(oldSvc.Spec, newSvc.Spec) ||
		!equality.Semantic.DeepEqual(oldSvc.Labels, newSvc.Labels) ||
		!equality.Semantic.DeepEqual(oldSvc.Annotations, newSvc.Annotations) ||
		!equality.Semantic.DeepEqual(oldSvc.OwnerReferences, newSvc.OwnerReferences)
}

func (r *PiHoleClusterReconciler) watchActivePiholeConfig(
	ctx context.Context,
	obj client.Object,
) []reconcile.Request {
	config, ok := obj.(*piholev1alpha1.PiHoleConfig)

	if !ok {
		return nil
	}

	if !meta.IsStatusConditionTrue(
		config.Status.Conditions,
		status.ConditionConfigReady,
	) {
		return nil
	}

	clusterName := config.Spec.ClusterRef.Name
	if clusterName == "" {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      config.Name,
				Namespace: config.Namespace,
			},
		},
	}
}
