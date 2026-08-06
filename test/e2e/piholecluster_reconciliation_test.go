package e2e

// Minimal Cluster
// Highly customized cluster (with DHCP)

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/test/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This file assumes the generated e2e_test.go suite has already:
//   - created the Kind cluster;
//   - built and loaded the manager image;
//   - installed the CRDs;
//   - deployed the controller manager.
//
// Keep the same e2e build tag strategy used by the generated suite and Makefile.
var _ = Describe("PiHoleCluster reconciliation", func() {
	const (
		clusterName     = "e2e-test-cluster"
		piholeImage     = "pihole/pihole:2026.05.0"
		adminSecretName = "pihole-admin-password"
		namespace       = "pihole-ha-operator-system"
		targetReplicas  = 3
	)

	var cluster v1alpha1.PiHoleCluster

	BeforeEach(func() {
		By("creating the Pi-hole admin password Secret")

		cmd := exec.Command(
			"kubectl",
			"create",
			"secret",
			"generic",
			adminSecretName,
			"--namespace",
			namespace,
			"--from-literal=password=admin",
		)

		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())

		By("Creating minimal cluster")
		cluster = v1alpha1.PiHoleCluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "pihole.paldab.nl/v1alpha1",
				Kind:       "PiHoleCluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: v1alpha1.PiHoleClusterSpec{
				Image:                       piholeImage,
				Replicas:                    new(int32(targetReplicas)),
				ExistingAdminPasswordSecret: adminSecretName,
			},
		}

		utils.CreatePiholeCluster(&cluster, namespace)
	})

	AfterEach(func() {
		By("deleting the test PiHoleCluster and Secret")

		cmd := exec.Command(
			"kubectl",
			"delete",
			"piholecluster",
			clusterName,
			"--namespace",
			namespace,
			"--ignore-not-found=true",
		)

		_, _ = utils.Run(cmd)

		cmd = exec.Command(
			"kubectl",
			"delete",
			"secret",
			adminSecretName,
			"--namespace",
			namespace,
			"--ignore-not-found=true",
		)

		_, _ = utils.Run(cmd)
	})

	Context("PiHoleCluster reconciliation", func() {
		It("ensuring statefulset of piholecluster have been created with the correct configuration", func() {
			By("verifying that statefulset has been created and has the right ready replicas")
			Eventually(func() error {
				expectedReplicas := int(*cluster.Spec.Replicas)
				stsSucceeded := utils.AssertStatefulsetCreatedSuccessfully(&cluster, expectedReplicas)
				Expect(stsSucceeded).To(BeTrue())

				return nil

			}, 2*time.Minute, 2*time.Second).Should(Succeed())
		})

		It("creates the DNS and web Services with the expected configuration", func() {
			expectedSelector := map[string]string{
				"paldab.nl/cluster":      clusterName,
				"paldab.nl/instanceRole": "primary",
			}

			serviceNames := []string{
				clusterName + "-dns",
				clusterName + "-web",
			}

			for _, serviceName := range serviceNames {
				By(fmt.Sprintf("verifying Service %q", serviceName))

				Eventually(func() error {
					cmd := exec.Command(
						"kubectl",
						"get",
						"service",
						serviceName,
						"--namespace",
						namespace,
						"--output",
						"json",
					)

					output, err := utils.Run(cmd)
					Expect(err).NotTo(HaveOccurred())

					var service corev1.Service
					err = json.Unmarshal([]byte(output), &service)
					Expect(err).NotTo(HaveOccurred())

					Expect(service.Name).To(Equal(serviceName))
					Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))

					for label, expectedValue := range expectedSelector {
						Expect(service.Spec.Selector).To(HaveKeyWithValue(label, expectedValue),
							"Service %q has an unexpected selector",
							serviceName,
						)
					}

					return err
				}, time.Minute, 2*time.Second).Should(Succeed())
			}
		})

		It("should handle cluster mutations", func() {
			By("verifying that statefulset has been created and has the right ready replicas")
			Eventually(func() error {
				expectedReplicas := int(*cluster.Spec.Replicas)
				stsSucceeded := utils.AssertStatefulsetCreatedSuccessfully(&cluster, expectedReplicas)
				Expect(stsSucceeded).To(BeTrue())

				return nil

			}, 2*time.Minute, 2*time.Second).Should(Succeed())

			By("changing the replicas on the cluster object to 5")
			newDesiredreplicas := 5

			Eventually(func() error {
				_, err := utils.PatchPiholeCluster(&cluster, fmt.Sprintf(`{"spec":{"replicas": %d}}`, newDesiredreplicas))

				Expect(err).NotTo(HaveOccurred())

				stsHasDesiredReplicasAfterPatch := utils.AssertStatefulDesiredReplicas(&cluster, newDesiredreplicas)
				Expect(stsHasDesiredReplicasAfterPatch).To(BeTrue())

				stsHasReadyReplicasAfterPatch := utils.AssertStatefulReadyReplicas(&cluster, newDesiredreplicas)
				Expect(stsHasReadyReplicasAfterPatch).To(BeTrue())

				return err

			}, 2*time.Minute, 2*time.Second).Should(
				Succeed(), fmt.Sprintf("failed to patch cluster replicas to %d", newDesiredreplicas),
			)

			By("changing the replicas on the cluster object down to 2")
			newDesiredreplicas = 2

			Eventually(func() error {
				_, err := utils.PatchPiholeCluster(&cluster, fmt.Sprintf(`{"spec":{"replicas": %d}}`, newDesiredreplicas))

				Expect(err).NotTo(HaveOccurred())
				stsHasDesiredReplicasAfterPatch := utils.AssertStatefulDesiredReplicas(&cluster, newDesiredreplicas)
				Expect(stsHasDesiredReplicasAfterPatch).To(BeTrue())

				stsHasReadyReplicasAfterPatch := utils.AssertStatefulReadyReplicas(&cluster, newDesiredreplicas)
				Expect(stsHasReadyReplicasAfterPatch).To(BeTrue())

				return err
			}, 2*time.Minute, 2*time.Second).Should(
				Succeed(), fmt.Sprintf("failed to patch cluster replicas to %d", newDesiredreplicas),
			)
		})

		It("should correctly delete all sub resources of pihole cluster", func() {
			By("verifying that statefulset has been created")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "statefulset", clusterName, "--namespace", namespace)

				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred())
				return err
			}, time.Minute, 2*time.Second).Should(Succeed())

			utils.DeletePiholeCluster(&cluster, namespace)

			By("verifying the PiHoleCluster was deleted")

			Eventually(func() (string, error) {
				cmd := exec.Command(
					"kubectl",
					"get",
					"piholecluster",
					cluster.Name,
					"--namespace",
					cluster.Namespace,
					"--ignore-not-found",
					"-o",
					"name",
				)

				output, err := utils.Run(cmd)
				return strings.TrimSpace(output), err
			}, time.Minute, 2*time.Second).Should(BeEmpty())

			By("verifying the statefulset was deleted")
			Eventually(func() error {
				cmd := exec.Command(
					"kubectl",
					"get",
					"statefulset",
					cluster.Name,
					"-n",
					cluster.Namespace,
					"--ignore-not-found",
					"-o",
					"name",
				)
				output, err := utils.Run(cmd)

				Expect(err).NotTo(HaveOccurred())
				Expect(strings.TrimSpace(output)).To(BeEmpty())

				return err
			}).Should(Succeed())
		})
	})
})
