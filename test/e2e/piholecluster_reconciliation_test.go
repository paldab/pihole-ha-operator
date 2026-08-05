package e2e

// Minimal Cluster
// Highly customized cluster (with DHCP)

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/test/utils"
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
				Replicas:                    new(int32(3)),
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

	Context("PiHoleCLuster reconciliation", func() {
		It("ensuring subresources of piholecluster have been created", func() {
			By("verifying that statefulset has been created")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "statefulset", clusterName, "--namespace", namespace)

				_, err := utils.Run(cmd)
				Expect(err).NotTo(HaveOccurred())
				return err
			}, 1*time.Minute, 2*time.Second).Should(Succeed())

			By("verifying that statefulset configuration is correct")
			Eventually(func() (string, error) {
				cmd := exec.Command(
					"kubectl",
					"get",
					"statefulset",
					clusterName,
					"-n",
					namespace,
					"-o",
					"jsonpath={.spec.replicas}",
				)

				out, err := utils.Run(cmd)

				Expect(err).NotTo(HaveOccurred())

				return out, err
			}).Should(Equal(strconv.Itoa(int(*cluster.Spec.Replicas))))

			It("should correctly delete all sub resources of pihole cluster", func() {
				By("verifying that statefulset has been created")
				Eventually(func() error {
					cmd := exec.Command("kubectl", "get", "statefulset", clusterName, "--namespace", namespace)

					_, err := utils.Run(cmd)
					Expect(err).NotTo(HaveOccurred())
					return err
				}, 1*time.Minute, 2*time.Second).Should(Succeed())

				utils.DeletePiholeCluster(&cluster, namespace)

				By("verifying the PiHoleCluster was deleted")

				Eventually(func() (string, error) {
					cmd := exec.Command(
						"kubectl",
						"get",
						"piholecluster",
						clusterName,
						"--namespace",
						namespace,
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
						namespace,
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
})
