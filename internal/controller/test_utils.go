package controller

import (
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createMinimalCluster(typeNamespacedName types.NamespacedName) *v1alpha1.PiHoleCluster {
	return &v1alpha1.PiHoleCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      typeNamespacedName.Name,
			Namespace: typeNamespacedName.Namespace,
		},

		Spec: v1alpha1.PiHoleClusterSpec{
			ExistingAdminPasswordSecret: "test-admin-password",
			Image:                       "pihole/pihole:latest",
		},
	}
}

func createMinimalConfig(typeNamespacedName types.NamespacedName, clusterName string) *v1alpha1.PiHoleConfig {
	return &v1alpha1.PiHoleConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      typeNamespacedName.Name,
			Namespace: typeNamespacedName.Namespace,
		},

		Spec: v1alpha1.PiHoleConfigSpec{
			ClusterRef: v1alpha1.ClusterRef{
				Name: clusterName,
			},
		},
	}
}
