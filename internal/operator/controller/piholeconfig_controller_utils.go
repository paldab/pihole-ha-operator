package controller

import (
	"context"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *PiHoleConfigReconciler) mapConfigMapToPiHoleConfigs(
	ctx context.Context,
	obj client.Object,
) []reconcile.Request {
	clusterName := obj.GetLabels()[defaults.ClusterNameLabel]
	if clusterName == "" {
		return nil
	}

	var configs piholev1alpha1.PiHoleConfigList

	if err := r.List(
		ctx,
		&configs,
		client.InNamespace(obj.GetNamespace()),
		client.MatchingFields{
			clusterRefField: clusterName,
		},
	); err != nil {
		return nil
	}

	requests := make([]reconcile.Request, 0, len(configs.Items))

	for _, config := range configs.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(&config),
		})
	}

	return requests
}

const clusterRefField = ".spec.clusterRef.name"

func (r *PiHoleConfigReconciler) configsForCluster(
	ctx context.Context,
	namespace string,
	clusterName string,
) ([]piholev1alpha1.PiHoleConfig, error) {
	var configs piholev1alpha1.PiHoleConfigList

	if err := r.List(
		ctx,
		&configs,
		client.InNamespace(namespace),
		client.MatchingFields{
			clusterRefField: clusterName,
		},
	); err != nil {
		return nil, err
	}

	return configs.Items, nil
}

func activeConfig(configs []piholev1alpha1.PiHoleConfig) *piholev1alpha1.PiHoleConfig {
	if len(configs) == 0 {
		return nil
	}

	winner := &configs[0]

	for i := 1; i < len(configs); i++ {
		candidate := &configs[i]

		candidateOlder := candidate.CreationTimestamp.Time.Before(
			winner.CreationTimestamp.Time,
		)

		sameTimestamp := candidate.CreationTimestamp.Time.Equal(
			winner.CreationTimestamp.Time,
		)

		if candidateOlder ||
			(sameTimestamp && candidate.Name < winner.Name) {
			winner = candidate
		}
	}

	return winner
}
