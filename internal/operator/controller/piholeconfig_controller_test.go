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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
)

var _ = Describe("PiHoleConfig Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"
		)
		var clusterName = fmt.Sprintf("%s-cluster", resourceName)

		ctx := context.Background()

		typeNamespacedNameCluster := types.NamespacedName{
			Name:      clusterName,
			Namespace: resourceNamespace,
		}

		typeNamespacedNameConfig := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}

		piholecluster := piholev1alpha1.PiHoleCluster{}
		piholeconfig := piholev1alpha1.PiHoleConfig{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind PiHoleCluster")
			err := k8sClient.Get(ctx, typeNamespacedNameCluster, &piholecluster)
			if err != nil && errors.IsNotFound(err) {
				resource := createMinimalCluster(typeNamespacedNameConfig)
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating the custom resource for the Kind PiHoleConfig")
			err = k8sClient.Get(ctx, typeNamespacedNameConfig, &piholeconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := createMinimalConfig(typeNamespacedNameConfig, clusterName)
				fmt.Println(typeNamespacedNameConfig, piholecluster.Name)
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &piholev1alpha1.PiHoleConfig{}
			err := k8sClient.Get(ctx, typeNamespacedNameConfig, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance PiHoleConfig")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &PiHoleConfigReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedNameConfig,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
