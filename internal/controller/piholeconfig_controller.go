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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/resources"
)

// PiHoleConfigReconciler reconciles a PiHoleConfig object
type PiHoleConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pihole.paldab.nl,resources=piholeconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PiHoleConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *PiHoleConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var piholeConfig piholev1alpha1.PiHoleConfig
	if err := r.Get(ctx, req.NamespacedName, &piholeConfig); err != nil {
		log.Error(err, "could not find requested pihole config!", "piholeconfig", req.Name)
		return ctrl.Result{}, err
	}

	var piholeCluster piholev1alpha1.PiHoleCluster
	requestedCluster := types.NamespacedName{Name: piholeConfig.Spec.ClusterRef.Name, Namespace: piholeConfig.Namespace}
	if err := r.Get(ctx, requestedCluster, &piholeCluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("waiting for PiHoleCluster",
				"cluster", requestedCluster.Name,
				"namespace", requestedCluster.Namespace,
				"config", piholeConfig.Name,
			)
			return ctrl.Result{}, nil
		}

		log.Error(err, "could not find the pihole cluster of the requested pihole config", "cluster", piholeConfig.Spec.ClusterRef.Name, "namespace", piholeConfig.Namespace, "config", piholeConfig.Name)
		return ctrl.Result{}, err
	}

	log.Info("Reconciling PiHoleConfig",
		"namespace", req.Namespace,
		"name", req.Name,
		"cluster", piholeCluster.Name,
	)

	configCopy := piholeConfig.DeepCopy()
	defaults.ApplyDefaultConfigValues(configCopy, &piholeCluster)

	resourceContext := resources.ResourceContext{
		Ctx:       ctx,
		K8sClient: r.Client,
		Cluster:   &piholeCluster,
		Scheme:    r.Scheme,
	}

	if err := resources.EnsureAdListConfigmap(&resourceContext, configCopy); err != nil {
		log.Error(err, "something went wrong when ensuring the adlist configmap", "config", piholeConfig.Name)
		return ctrl.Result{}, err
	}

	if err := resources.EnsureCNAMEConfigmap(&resourceContext, configCopy); err != nil {
		log.Error(err, "something went wrong when ensuring the CNAME configmap", "config", piholeConfig.Name)
		return ctrl.Result{}, err
	}

	if err := resources.EnsureAdditionalHostsConfigmap(&resourceContext, configCopy); err != nil {
		log.Error(err, "something went wrong when ensuring the Additional Hosts configmap", "config", piholeConfig.Name)
		return ctrl.Result{}, err
	}

	if err := resources.EnsureCustomConfigMap(&resourceContext, configCopy); err != nil {
		log.Error(err, "something went wrong when ensuring the Custom Settings configmap", "config", piholeConfig.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PiHoleConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&piholev1alpha1.PiHoleConfig{}).
		Named("piholeconfig").
		Complete(r)
}
