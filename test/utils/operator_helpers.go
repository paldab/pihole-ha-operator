package utils

//nolint:staticcheck
import (
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func DeployOperator(managerImage string) {
	cmd := exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", managerImage))
	_, err := Run(cmd)
	Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")
}

func GetNameOfDeployedOperatorPod(g Gomega, namespace string) string {
	cmd := exec.Command("kubectl", "get",
		"pods", "-l", "control-plane=controller-manager",
		"-o", "go-template={{ range .items }}"+
			"{{ if not .metadata.deletionTimestamp }}"+
			"{{ .metadata.name }}"+
			"{{ \"\\n\" }}{{ end }}{{ end }}",
		"-n", namespace,
	)

	podOutput, err := Run(cmd)
	g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
	podNames := GetNonEmptyLines(podOutput)
	g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
	controllerPodName := podNames[0]
	g.Expect(controllerPodName).To(ContainSubstring("controller-manager"))

	return controllerPodName
}

func WaitForOperatorReady(operatorPodName, namespace string) {
	Eventually(func() error {
		cmd := exec.Command(
			"kubectl",
			"rollout",
			"status",
			"deployment",
			operatorPodName,
			"-n",
			namespace,
			"--timeout=10s",
		)

		_, err := Run(cmd)
		return err
	}, 2*time.Minute, 2*time.Second).Should(Succeed())
}

func CheckOperatorFailures(controllerPodName, namespace string) {
	By("Fetching controller manager pod logs")
	cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
	controllerLogs, err := Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
	} else {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
	}

	By("Fetching Kubernetes events")
	cmd = exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
	eventsOutput, err := Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
	} else {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
	}

	By("Fetching curl-metrics logs")
	cmd = exec.Command("kubectl", "logs", "curl-metrics", "-n", namespace)
	metricsOutput, err := Run(cmd)
	if err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Metrics logs:\n %s", metricsOutput)
	} else {
		_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get curl-metrics logs: %s", err)
	}

	By("Fetching controller manager pod description")
	cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
	podDescription, err := Run(cmd)
	if err == nil {
		fmt.Println("Pod description:\n", podDescription)
	} else {
		fmt.Println("Failed to describe controller pod")
	}
}
