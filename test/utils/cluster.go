package utils

//nolint:staticcheck
import (
	"bytes"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"sigs.k8s.io/yaml"
)

func CreatePiholeCluster(cluster *v1alpha1.PiHoleCluster, namespace string) {
	manifest, err := yaml.Marshal(cluster)
	Expect(err).NotTo(HaveOccurred())

	By("applying the PiHoleCluster")

	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(manifest)

	output, err := Run(cmd)
	Expect(err).NotTo(
		HaveOccurred(),
		"failed to apply PiHoleCluster: %s",
		output,
	)

	By("verifying that the PiHoleCluster exists")
	Eventually(func() error {
		cmd := exec.Command(
			"kubectl",
			"get",
			"piholecluster",
			cluster.Name,
			"--namespace",
			namespace,
		)

		_, err := Run(cmd)
		return err
	}, 1*time.Minute, 2*time.Second).Should(Succeed())
}

func DeletePiholeCluster(cluster *v1alpha1.PiHoleCluster, namespace string) {
	By("deleting the PiHoleCluster")

	cmd := exec.Command("kubectl", "delete", "piholecluster", cluster.Name, "--namespace", namespace)

	_, err := Run(cmd)
	Expect(err).NotTo(
		HaveOccurred(),
		"failed to delete PiHoleCluster: %s",
	)
}

func PatchPiholeCluster(cluster *v1alpha1.PiHoleCluster, patchSpec string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"patch",
		"piholecluster",
		cluster.Name,
		"-n",
		cluster.Namespace,
		"--type=merge",
		"-p",
		patchSpec,
	)

	return Run(cmd)
}
