package utils

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
)

func AssertStatefulsetCreatedSuccessfully(cluster *v1alpha1.PiHoleCluster, expectedReplicas int) bool {
	succeeded := assertStatefulsetExists(cluster)

	if !succeeded {
		return succeeded
	}

	succeeded = AssertStatefulDesiredReplicas(cluster, expectedReplicas)

	if !succeeded {
		return succeeded
	}

	succeeded = AssertStatefulReadyReplicas(cluster, expectedReplicas)

	if !succeeded {
		return succeeded
	}

	return true
}

func assertStatefulsetExists(cluster *v1alpha1.PiHoleCluster) bool {
	By("verifying that statefulset has been created")
	return Eventually(func() error {
		cmd := exec.Command("kubectl", "get", "statefulset", cluster.Name, "--namespace", cluster.Namespace)

		_, err := Run(cmd)

		Expect(err).NotTo(HaveOccurred())

		return err
	}, time.Minute, 2*time.Second).Should(Succeed())
}

func AssertStatefulDesiredReplicas(cluster *v1alpha1.PiHoleCluster, expectedReplicas int) bool {
	By("verifying that statefulset configuration is correct by checking the if the replicas on cluster match the statefulset")

	return Eventually(func() (string, error) {
		cmd := exec.Command(
			"kubectl",
			"get",
			"statefulset",
			cluster.Name,
			"-n",
			cluster.Namespace,
			"-o",
			"jsonpath={.spec.replicas}",
		)

		out, err := Run(cmd)

		Expect(err).NotTo(HaveOccurred())

		return out, err
	}).Should(Equal(strconv.Itoa(expectedReplicas)))
}

func AssertStatefulReadyReplicas(cluster *v1alpha1.PiHoleCluster, expectedReplicas int) bool {
	By("waiting for all StatefulSet replicas to become ready")

	readinessSecondsPerPod := 40 * time.Second
	podReadinessTimeout := time.Duration(max(expectedReplicas, 1)) * readinessSecondsPerPod

	return Eventually(func() (int, error) {
		cmd := exec.Command(
			"kubectl",
			"get",
			"statefulset",
			cluster.Name,
			"--namespace",
			cluster.Namespace,
			"--output",
			"jsonpath={.status.readyReplicas}",
		)

		output, err := Run(cmd)

		if err != nil {
			return 0, err
		}

		output = strings.TrimSpace(output)

		readyReplicas, err := strconv.Atoi(output)
		if err != nil {
			return 0, fmt.Errorf(
				"invalid StatefulSet readyReplicas value %q: %w",
				output,
				err,
			)
		}

		return readyReplicas, nil
	}, podReadinessTimeout, 2*time.Second).Should(
		Equal(expectedReplicas), "expected all StatefulSet replicas to become ready",
	)
}
