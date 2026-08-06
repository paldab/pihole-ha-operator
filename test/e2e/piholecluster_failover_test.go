package e2e

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PiHoleCluster failover tests", func() {
	const (
		clusterName     = "e2e-pihole-cluster-failover"
		piholeImage     = "pihole/pihole:2026.05.0"
		adminSecretName = "pihole-admin-password"
		namespace       = "pihole-ha-operator-system"
		targetReplicas  = 3
	)
	var cluster v1alpha1.PiHoleCluster
	// TODO
	// failover with 2 replicas
	// failover with 3 replicas

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

})
