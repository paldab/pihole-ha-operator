// Package resources ensures that the resources are created with the correct data
// This package interacts with the k8s api client
package resources

import (
	"context"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceContext struct {
	Ctx       context.Context
	K8sClient client.Client
	Cluster   *v1alpha1.PiHoleCluster
	Scheme    *runtime.Scheme
}
