/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/resources"
	"github.com/paldab/pihole-ha-operator/internal/operator/status"
)

// PiHoleConfigReconciler reconciles a PiHoleConfig object
type PiHoleConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PiHoleConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var piholeConfig piholev1alpha1.PiHoleConfig
	if err := r.Get(ctx, req.NamespacedName, &piholeConfig); err != nil {
		// Object was deleted
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "could not find requested pihole config!", "config", req.Name)
		return ctrl.Result{}, err
	}

	clusterName := piholeConfig.Spec.ClusterRef.Name
	configs, err := r.configsForCluster(ctx, piholeConfig.Namespace, clusterName)
	if err != nil {
		log.Error(err, "could not list PiholeConfigs for cluster", "cluster", clusterName, "config", piholeConfig.Name)
		return ctrl.Result{}, err
	}

	activeConfig := activeConfig(configs)

	// this case should not commonly happen because PiholeConfig shuld be indexed in the result
	// Don't write configuration if there is no ownership of PiholeConfig
	if activeConfig == nil {
		log.Info(
			"could not determine active PiholeConfig",
			"cluster", clusterName,
			"config", piholeConfig.Name,
		)

		return ctrl.Result{
			RequeueAfter: time.Second,
		}, nil
	}

	// Rejects duplicate configs for the same cluster
	if activeConfig.UID != piholeConfig.UID {
		if err := status.SetConfigReadyCondition(
			ctx,
			r.Client,
			&piholeConfig,
			metav1.ConditionFalse,
			status.ReasonDuplicateClusterRef,
			fmt.Sprintf(
				"PiHoleCluster %q is already managed by PiHoleConfig %q",
				clusterName,
				activeConfig.Name,
			),
		); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	var piholeCluster piholev1alpha1.PiHoleCluster
	requestedCluster := types.NamespacedName{Name: clusterName, Namespace: piholeConfig.Namespace}
	if err := r.Get(ctx, requestedCluster, &piholeCluster); err != nil {
		if apierrors.IsNotFound(err) {
			if statusUpdateErr := status.SetConfigReadyCondition(
				ctx,
				r.Client,
				&piholeConfig,
				metav1.ConditionFalse,
				status.ReasonClusterNotFound,
				fmt.Sprintf(
					"Failed to find PiholeCluster %s",
					clusterName,
				),
			); statusUpdateErr != nil {
				return ctrl.Result{}, statusUpdateErr
			}

			return ctrl.Result{}, nil
		}

		log.Error(
			err,
			"could not find the pihole cluster of the requested pihole config",
			"cluster", clusterName,
			"namespace", piholeConfig.Namespace,
			"config", piholeConfig.Name,
		)
		return ctrl.Result{}, err
	}

	configCopy := piholeConfig.DeepCopy()
	defaults.ApplyDefaultConfigValues(configCopy, &piholeCluster)

	resourceContext := resources.ResourceContext{
		Ctx:       ctx,
		K8sClient: r.Client,
		Cluster:   &piholeCluster,
		Scheme:    r.Scheme,
	}

	// Fill configmaps with data
	for component := range defaults.PiholeStaticMountConfig {
		if component == defaults.StsVolumeName {
			continue
		}

		if err := resources.CreateConfigmapWrapper(&resourceContext, configCopy, component); err != nil {
			log.Error(err, "something went wrong when ensuring configmap", "config", piholeConfig.Name, "type", string(component))

			if statusUpdateErr := status.SetConfigReadyCondition(
				ctx,
				r.Client,
				&piholeConfig,
				metav1.ConditionFalse,
				status.ReasonReconcileFailed,
				fmt.Sprintf(
					"Failed to reconcile %s configuration",
					component,
				),
			); statusUpdateErr != nil {
				return ctrl.Result{}, statusUpdateErr
			}

			return ctrl.Result{}, err
		}
	}

	if err := status.SetConfigReadyCondition(
		ctx,
		r.Client,
		&piholeConfig,
		metav1.ConditionTrue,
		status.ReasonConfigurationApplied,
		fmt.Sprintf(
			"Configuration successfully applied to PiHoleCluster %q",
			piholeCluster.Name,
		),
	); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

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

// SetupWithManager sets up the controller with the Manager.
func (r *PiHoleConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&piholev1alpha1.PiHoleConfig{},
		clusterRefField,
		func(obj client.Object) []string {
			config := obj.(*piholev1alpha1.PiHoleConfig)

			if config.Spec.ClusterRef.Name == "" {
				return nil
			}

			return []string{config.Spec.ClusterRef.Name}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&piholev1alpha1.PiHoleConfig{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.mapConfigMapToPiHoleConfigs),
		).
		Named("piholeconfig").
		Complete(r)
}
