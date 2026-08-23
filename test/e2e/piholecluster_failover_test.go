package e2e

import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("PiHoleCluster failover tests", func() {
	const (
		clusterName     = "e2e-pihole-cluster-failover"
		piholeImage     = "pihole/pihole:2026.05.0"
		adminSecretName = "pihole-admin-password"
		namespace       = "pihole-ha-operator-system"
	)
	var cluster v1alpha1.PiHoleCluster

	var targetReplicas int32

	JustBeforeEach(func() {
		By(fmt.Sprintf(
			"creating PiHoleCluster with %d replicas",
			targetReplicas,
		))

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
				Image:    piholeImage,
				Replicas: &targetReplicas,
				ExistingSecretRef: v1alpha1.ExistingPasswordSecretRef{
					SecretName: adminSecretName,
				},
			},
		}

		utils.CreatePiholeCluster(&cluster, namespace)

		By("waiting for the cluster to become ready")
		utils.WaitForClusterReady(namespace, clusterName)

		By("waiting for exactly one primary")
		utils.WaitForExactlyOnePrimary(namespace, clusterName)
	})

	Context("with 1 replica", func() {
		BeforeEach(func() {
			targetReplicas = 1
		})

		It("starts with exactly one primary", func() {
			primary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			Expect(primary.Name).NotTo(BeEmpty())

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(BeEmpty())

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				1,
			)
		})

		It("recovers when the only primary is deleted", func() {
			oldPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			oldUID := oldPrimary.UID

			By("deleting the only primary")
			utils.DeletePod(namespace, oldPrimary.Name)

			By("waiting for the StatefulSet to recreate the pod")
			recreated := utils.WaitForPodUIDChange(
				namespace,
				oldPrimary.Name,
				oldUID,
			)

			Expect(recreated.Name).To(Equal(oldPrimary.Name))
			Expect(recreated.UID).NotTo(Equal(oldUID))

			By("waiting for the recreated pod to become Ready")
			utils.WaitForPodReady(
				namespace,
				oldPrimary.Name,
			)

			By("waiting for it to become the single primary again")
			newPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			Expect(newPrimary.Name).To(Equal(oldPrimary.Name))

			utils.WaitForClusterReady(namespace, clusterName)
			utils.WaitForFailoverIdle(namespace, clusterName)

			utils.WaitForPrimaryServicesToTarget(
				namespace,
				clusterName,
				newPrimary.Name,
			)
		})

		It("recovers from repeated primary deletions without stale state", func() {
			currentPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			for i := range 3 {
				By(fmt.Sprintf(
					"deleting the only primary - iteration %d",
					i+1,
				))

				oldUID := currentPrimary.UID
				podName := currentPrimary.Name

				utils.DeletePod(namespace, podName)

				recreated := utils.WaitForPodUIDChange(
					namespace,
					podName,
					oldUID,
				)

				utils.WaitForPodReady(
					namespace,
					recreated.Name,
				)

				currentPrimary = utils.WaitForExactlyOnePrimary(
					namespace,
					clusterName,
				)

				Expect(currentPrimary.Name).To(Equal(podName))
				Expect(currentPrimary.UID).NotTo(Equal(oldUID))

				utils.WaitForClusterReady(
					namespace,
					clusterName,
				)

				utils.WaitForFailoverIdle(
					namespace,
					clusterName,
				)
			}
		})
	})

	Context("with 2 replicas", func() {
		BeforeEach(func() {
			targetReplicas = 2
		})

		It("starts with one primary and one standby", func() {
			utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(HaveLen(1))

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				2,
			)
		})

		It("does not fail over when a standby is deleted", func() {
			primary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(HaveLen(1))

			standby := standbys[0]
			oldStandbyUID := standby.UID

			By("deleting the standby")
			utils.DeletePod(namespace, standby.Name)

			By("ensuring the current primary does not move")
			utils.ExpectPrimaryStays(
				namespace,
				clusterName,
				primary.Name,
				10*time.Second,
			)

			By("waiting for the standby to be recreated")
			utils.WaitForPodUIDChange(
				namespace,
				standby.Name,
				oldStandbyUID,
			)

			utils.WaitForPodReady(
				namespace,
				standby.Name,
			)

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				2,
			)

			currentPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			Expect(currentPrimary.Name).To(Equal(primary.Name))

			utils.WaitForFailoverIdle(namespace, clusterName)
		})

		It("promotes the standby when the primary is deleted", func() {
			oldPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(HaveLen(1))

			expectedReplacement := standbys[0]
			oldPrimaryUID := oldPrimary.UID

			By("deleting the primary")
			utils.DeletePod(namespace, oldPrimary.Name)

			By("waiting for the standby to become primary")
			newPrimary := utils.WaitForPrimaryChange(
				namespace,
				clusterName,
				oldPrimary.Name,
			)

			Expect(newPrimary.Name).To(Equal(expectedReplacement.Name))

			By("waiting for the old primary StatefulSet pod to return")
			utils.WaitForPodUIDChange(
				namespace,
				oldPrimary.Name,
				oldPrimaryUID,
			)

			utils.WaitForPodReady(
				namespace,
				oldPrimary.Name,
			)

			By("ensuring the recovered old primary does not steal leadership")
			utils.ExpectPrimaryStays(
				namespace,
				clusterName,
				newPrimary.Name,
				10*time.Second,
			)

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				2,
			)

			utils.WaitForClusterReady(namespace, clusterName)
			utils.WaitForFailoverIdle(namespace, clusterName)

			utils.WaitForPrimaryServicesToTarget(
				namespace,
				clusterName,
				newPrimary.Name,
			)
		})

		It("can fail over repeatedly between replicas", func() {
			currentPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			for i := range 3 {
				oldPrimary := currentPrimary
				oldUID := oldPrimary.UID

				By(fmt.Sprintf(
					"deleting current primary %s - iteration %d",
					oldPrimary.Name,
					i+1,
				))

				utils.DeletePod(
					namespace,
					oldPrimary.Name,
				)

				currentPrimary = utils.WaitForPrimaryChange(
					namespace,
					clusterName,
					oldPrimary.Name,
				)

				Expect(currentPrimary.Name).
					NotTo(Equal(oldPrimary.Name))

				utils.WaitForPodUIDChange(
					namespace,
					oldPrimary.Name,
					oldUID,
				)

				utils.WaitForPodReady(
					namespace,
					oldPrimary.Name,
				)

				utils.WaitForReadyReplicas(
					namespace,
					clusterName,
					2,
				)

				utils.WaitForClusterReady(
					namespace,
					clusterName,
				)

				utils.ExpectPrimaryStays(
					namespace,
					clusterName,
					currentPrimary.Name,
					5*time.Second,
				)
			}
		})

		It("recovers after both replicas disappear", func() {
			By("deleting all Pi-hole pods")
			utils.DeleteAllClusterPods(
				namespace,
				clusterName,
			)

			By("verifying recovery never settles into split brain")
			utils.ExpectAtMostOnePrimaryFor(
				namespace,
				clusterName,
				15*time.Second,
			)

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				2,
			)

			primary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			utils.WaitForClusterReady(namespace, clusterName)
			utils.WaitForFailoverIdle(namespace, clusterName)

			utils.WaitForPrimaryServicesToTarget(
				namespace,
				clusterName,
				primary.Name,
			)
		})
	})

	Context("with 3 replicas", func() {
		BeforeEach(func() {
			targetReplicas = 3
		})

		It("starts with one primary and two standbys", func() {
			utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(HaveLen(2))

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				3,
			)
		})

		It("does not fail over when one standby is deleted", func() {
			primary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(HaveLen(2))

			standby := standbys[0]
			oldUID := standby.UID

			utils.DeletePod(
				namespace,
				standby.Name,
			)

			utils.ExpectPrimaryStays(
				namespace,
				clusterName,
				primary.Name,
				10*time.Second,
			)

			utils.WaitForPodUIDChange(
				namespace,
				standby.Name,
				oldUID,
			)

			utils.WaitForPodReady(
				namespace,
				standby.Name,
			)

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				3,
			)

			currentPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			Expect(currentPrimary.Name).
				To(Equal(primary.Name))
		})

		It("does not fail over when both standbys are deleted", func() {
			primary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			standbys, err := utils.GetStandbyPods(
				namespace,
				clusterName,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(standbys).To(HaveLen(2))

			oldUIDs := map[string]types.UID{}

			for _, standby := range standbys {
				oldUIDs[standby.Name] = standby.UID

				utils.DeletePod(
					namespace,
					standby.Name,
				)
			}

			utils.ExpectPrimaryStays(
				namespace,
				clusterName,
				primary.Name,
				10*time.Second,
			)

			for _, standby := range standbys {
				utils.WaitForPodUIDChange(
					namespace,
					standby.Name,
					oldUIDs[standby.Name],
				)

				utils.WaitForPodReady(
					namespace,
					standby.Name,
				)
			}

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				3,
			)

			currentPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			Expect(currentPrimary.Name).
				To(Equal(primary.Name))

			utils.WaitForFailoverIdle(
				namespace,
				clusterName,
			)
		})

		It("promotes one standby when the primary is deleted", func() {
			oldPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			oldUID := oldPrimary.UID

			By("deleting the current primary")
			utils.DeletePod(
				namespace,
				oldPrimary.Name,
			)

			By("waiting for one of the standbys to be promoted")
			newPrimary := utils.WaitForPrimaryChange(
				namespace,
				clusterName,
				oldPrimary.Name,
			)

			Expect(newPrimary.Name).
				NotTo(Equal(oldPrimary.Name))

			By("waiting for the old primary pod to return")
			utils.WaitForPodUIDChange(
				namespace,
				oldPrimary.Name,
				oldUID,
			)

			utils.WaitForPodReady(
				namespace,
				oldPrimary.Name,
			)

			By("ensuring the returned pod does not reclaim primary")
			utils.ExpectPrimaryStays(
				namespace,
				clusterName,
				newPrimary.Name,
				10*time.Second,
			)

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				3,
			)

			utils.WaitForClusterReady(namespace, clusterName)
			utils.WaitForFailoverIdle(namespace, clusterName)

			utils.WaitForPrimaryServicesToTarget(
				namespace,
				clusterName,
				newPrimary.Name,
			)
		})

		It("survives repeated sequential primary failures", func() {
			currentPrimary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			for i := range 4 {
				oldPrimary := currentPrimary
				oldUID := oldPrimary.UID

				By(fmt.Sprintf(
					"deleting primary %s - failover %d",
					oldPrimary.Name,
					i+1,
				))

				utils.DeletePod(
					namespace,
					oldPrimary.Name,
				)

				currentPrimary = utils.WaitForPrimaryChange(
					namespace,
					clusterName,
					oldPrimary.Name,
				)

				Expect(currentPrimary.Name).
					NotTo(Equal(oldPrimary.Name))

				utils.WaitForPodUIDChange(
					namespace,
					oldPrimary.Name,
					oldUID,
				)

				utils.WaitForPodReady(
					namespace,
					oldPrimary.Name,
				)

				utils.WaitForReadyReplicas(
					namespace,
					clusterName,
					3,
				)

				utils.WaitForClusterReady(
					namespace,
					clusterName,
				)

				utils.WaitForFailoverIdle(
					namespace,
					clusterName,
				)

				utils.ExpectPrimaryStays(
					namespace,
					clusterName,
					currentPrimary.Name,
					5*time.Second,
				)

				utils.WaitForPrimaryServicesToTarget(
					namespace,
					clusterName,
					currentPrimary.Name,
				)
			}
		})

		It("recovers after all three Pi-hole pods disappear", func() {
			By("deleting every Pi-hole pod")
			utils.DeleteAllClusterPods(
				namespace,
				clusterName,
			)

			By("ensuring the recovering cluster never settles with multiple primaries")
			utils.ExpectAtMostOnePrimaryFor(
				namespace,
				clusterName,
				15*time.Second,
			)

			utils.WaitForReadyReplicas(
				namespace,
				clusterName,
				3,
			)

			primary := utils.WaitForExactlyOnePrimary(
				namespace,
				clusterName,
			)

			utils.WaitForClusterReady(namespace, clusterName)
			utils.WaitForFailoverIdle(namespace, clusterName)

			utils.WaitForPrimaryServicesToTarget(
				namespace,
				clusterName,
				primary.Name,
			)
		})
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

})
