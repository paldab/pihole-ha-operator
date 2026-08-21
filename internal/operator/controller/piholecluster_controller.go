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

// Package controller contains Kubernetes controllers for managing Pihole Clusters which create Pihole resources.
package controller

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/failover"
	"github.com/paldab/pihole-ha-operator/internal/operator/resources"
	"github.com/paldab/pihole-ha-operator/internal/operator/status"
)

// PiHoleClusterReconciler reconciles a PiHoleCluster object
type PiHoleClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

func (r *PiHoleClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var piholeCluster piholev1alpha1.PiHoleCluster

	if err := r.Get(ctx, req.NamespacedName, &piholeCluster); err != nil {
		// Object was deleted
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "could not find requested pihole cluster!", "piholecluster", req.Name)
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling PiHoleCluster",
		"namespace", req.Namespace,
		"name", req.Name,
	)

	clusterCopy := piholeCluster.DeepCopy()

	defaults.ApplyDefaultClusterValues(clusterCopy)

	resourceContext := resources.ResourceContext{
		Ctx:       ctx,
		K8sClient: r.Client,
		Cluster:   clusterCopy,
		Scheme:    r.Scheme,
	}

	// create empty configmaps
	if err := resources.CreateInitialEmptyPiholeConfigmaps(&resourceContext); err != nil {
		log.Error(err, "failed to create initial configmaps for pihole cluster", "cluster", clusterCopy.Name)
		return ctrl.Result{}, err
	}

	if err := resources.EnsureStatefulSet(&resourceContext); err != nil {
		log.Error(err, "failed to reconcile StatefulSet")
		return ctrl.Result{}, err
	}

	if err := resources.EnsureServices(&resourceContext); err != nil {
		log.Error(err, "failed to reconcile Services")
		return ctrl.Result{}, err
	}

	if err := resources.EnsureIngress(&resourceContext); err != nil {
		log.Error(err, "failed to reconcile Ingress")
		return ctrl.Result{}, err
	}

	managedSts := &appsv1.StatefulSet{}

	if err := r.Get(ctx, req.NamespacedName, managedSts); err != nil {
		log.Error(err, "failed to fetch managed Statefulset")
		return ctrl.Result{}, err
	}

	clusterOwnedPodLabels := defaults.PiholeOperatorLabels(piholeCluster.Name)
	clusterOwnedpods := &corev1.PodList{}
	if err := r.List(
		ctx,
		clusterOwnedpods,
		client.InNamespace(piholeCluster.Namespace),
		client.MatchingLabels(clusterOwnedPodLabels),
	); err != nil {
		log.Error(err, "failed to fetch managed Pihole Pods")
		return ctrl.Result{}, err
	}

	var failoverResult failover.FailoverResult
	var err error

	desiredReplicas := ptr.Deref(piholeCluster.Spec.Replicas, int32(1))

	if desiredReplicas == 1 && len(clusterOwnedpods.Items) == 1 {
		onlyPod := clusterOwnedpods.Items[0]
		failoverResult, err = failover.ReconcileFailoverSingleInstance(ctx, r.Client, &onlyPod)
	} else if desiredReplicas > 1 {
		failoverResult, err = failover.ReconcileFailoverMultiInstance(ctx, r.Client, clusterOwnedpods)
	}

	if err != nil {
		// handle startup cases where pods are still getting ready
		if failoverResult.Reason == failover.ReasonLeaderUnavailable {
			if err := status.UpdateClusterStatus(&resourceContext, managedSts, &failoverResult); err != nil {
				log.Error(err, "failed to update Pihole cluster status",
					"cluster", clusterCopy.Name,
					"failover_reason", failoverResult.Reason)
				return ctrl.Result{}, err
			}

			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		log.Error(err, "something went wrong during the failover process")
		return ctrl.Result{}, err
	}

	if err := status.UpdateClusterStatus(&resourceContext, managedSts, &failoverResult); err != nil {
		log.Error(err, "failed to update Pihole cluster status",
			"cluster", clusterCopy.Name,
			"leader", failoverResult.Leader.Name,
			"failover_reason", failoverResult.Reason)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PiHoleClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&piholev1alpha1.PiHoleCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1.Ingress{}).
		Named("piholecluster").
		Complete(r)
}
