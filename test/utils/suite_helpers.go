// Package utils contain utilities for e2e tests
package utils

//nolint:staticcheck
import (
	"bytes"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func CreateE2ETestNamespace(namespace string) {
	By("creating manager namespace")
	cmd := exec.Command("kubectl", "create", "ns", namespace)
	_, err := Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to create namespace")

	// By("labeling the namespace to enforce the restricted security policy")
	// cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
	// 	"pod-security.kubernetes.io/enforce=restricted")
	// _, err = Run(cmd)
	// Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")
}

func InstallCrds() {
	By("installing CRDs")
	cmd := exec.Command("make", "install")
	_, err := Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")
}

// CreateObjectIfNotExists expects a kubectl create string args like ("kubectl", "create", "clusterrolebinding", "test")
func CreateObjectIfNotExists(command string, args ...string) error {
	createCmd := exec.Command(command, args...)
	manifest, err := createCmd.Output()
	Expect(err).NotTo(HaveOccurred())

	applyCmd := exec.Command("kubectl", "apply", "-f", "-")
	applyCmd.Stdin = bytes.NewReader(manifest)

	_, err = Run(applyCmd)

	return err
}
