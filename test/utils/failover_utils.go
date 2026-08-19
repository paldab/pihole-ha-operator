package utils

//nolint:staticcheck
import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/gomega"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/status"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	PrimaryLabelKey   = defaults.RoleLabel
	PrimaryLabelValue = defaults.LeaderLabel

	DefaultE2ETimeout = 3 * time.Minute
	PollInterval      = 2 * time.Second
)

func kubectlJSON(out any, args ...string) error {
	cmd := exec.Command("kubectl", args...)

	output, err := Run(cmd)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(output), out)
}

func GetPiHoleCluster(
	namespace string,
	clusterName string,
) (piholev1alpha1.PiHoleCluster, error) {
	var cluster piholev1alpha1.PiHoleCluster

	err := kubectlJSON(
		&cluster,
		"get",
		"piholecluster",
		clusterName,
		"--namespace",
		namespace,
		"-o",
		"json",
	)

	return cluster, err
}

func WaitForClusterCondition(
	namespace string,
	clusterName string,
	conditionType string,
	expectedStatus metav1.ConditionStatus,
) metav1.Condition {
	var result metav1.Condition

	Eventually(func() error {
		cluster, err := GetPiHoleCluster(namespace, clusterName)
		if err != nil {
			return err
		}

		condition := meta.FindStatusCondition(
			cluster.Status.Conditions,
			conditionType,
		)

		if condition == nil {
			return fmt.Errorf(
				"condition %q not found",
				conditionType,
			)
		}

		if condition.Status != expectedStatus {
			return fmt.Errorf(
				"condition %q is %s, expected %s; reason=%s",
				conditionType,
				condition.Status,
				expectedStatus,
				condition.Reason,
			)
		}

		result = *condition
		return nil
	}, DefaultE2ETimeout, PollInterval).Should(Succeed())

	return result
}

func WaitForClusterReady(
	namespace string,
	clusterName string,
) {
	WaitForClusterCondition(
		namespace,
		clusterName,
		"Ready",
		metav1.ConditionTrue,
	)
}

func WaitForFailoverIdle(
	namespace string,
	clusterName string,
) {
	WaitForClusterCondition(
		namespace,
		clusterName,
		status.TypeFailingOver,
		metav1.ConditionFalse,
	)
}

func WaitForReadyReplicas(
	namespace string,
	clusterName string,
	expected int32,
) {
	Eventually(func() (int32, error) {
		cluster, err := GetPiHoleCluster(namespace, clusterName)
		if err != nil {
			return 0, err
		}

		return cluster.Status.ReadyReplicas, nil
	}, DefaultE2ETimeout, PollInterval).Should(Equal(expected))
}

// -----------------------------------------------------------------------------
// StatefulSet
// -----------------------------------------------------------------------------

func GetOwnedStatefulSet(
	namespace string,
	clusterName string,
) (appsv1.StatefulSet, error) {
	var list appsv1.StatefulSetList

	if err := kubectlJSON(
		&list,
		"get",
		"statefulsets",
		"--namespace",
		namespace,
		"-o",
		"json",
	); err != nil {
		return appsv1.StatefulSet{}, err
	}

	var matches []appsv1.StatefulSet

	for _, sts := range list.Items {
		for _, owner := range sts.OwnerReferences {
			if owner.Kind == "PiHoleCluster" &&
				owner.Name == clusterName {

				matches = append(matches, sts)
				break
			}
		}
	}

	if len(matches) != 1 {
		return appsv1.StatefulSet{}, fmt.Errorf(
			"expected exactly one StatefulSet owned by PiHoleCluster %q, got %d",
			clusterName,
			len(matches),
		)
	}

	return matches[0], nil
}

// -----------------------------------------------------------------------------
// Pods
// -----------------------------------------------------------------------------

func GetPod(
	namespace string,
	podName string,
) (corev1.Pod, error) {
	var pod corev1.Pod

	err := kubectlJSON(
		&pod,
		"get",
		"pod",
		podName,
		"--namespace",
		namespace,
		"-o",
		"json",
	)

	return pod, err
}

func GetClusterPods(
	namespace string,
	clusterName string,
) ([]corev1.Pod, error) {
	sts, err := GetOwnedStatefulSet(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	var list corev1.PodList

	if err := kubectlJSON(
		&list,
		"get",
		"pods",
		"--namespace",
		namespace,
		"-o",
		"json",
	); err != nil {
		return nil, err
	}

	pods := make([]corev1.Pod, 0)

	for _, pod := range list.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "StatefulSet" &&
				owner.Name == sts.Name {

				pods = append(pods, pod)
				break
			}
		}
	}

	return pods, nil
}

func GetPrimaryPods(
	namespace string,
	clusterName string,
) ([]corev1.Pod, error) {
	pods, err := GetClusterPods(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	primaries := make([]corev1.Pod, 0)

	for _, pod := range pods {
		if pod.Labels[PrimaryLabelKey] == PrimaryLabelValue {
			primaries = append(primaries, pod)
		}
	}

	return primaries, nil
}

func GetStandbyPods(
	namespace string,
	clusterName string,
) ([]corev1.Pod, error) {
	pods, err := GetClusterPods(namespace, clusterName)
	if err != nil {
		return nil, err
	}

	standbys := make([]corev1.Pod, 0)

	for _, pod := range pods {
		if pod.Labels[PrimaryLabelKey] != PrimaryLabelValue {
			standbys = append(standbys, pod)
		}
	}

	return standbys, nil
}

func WaitForExactlyOnePrimary(
	namespace string,
	clusterName string,
) corev1.Pod {
	var primary corev1.Pod

	Eventually(func() error {
		primaries, err := GetPrimaryPods(namespace, clusterName)
		if err != nil {
			return err
		}

		if len(primaries) != 1 {
			return fmt.Errorf(
				"expected exactly one primary, got %d",
				len(primaries),
			)
		}

		primary = primaries[0]
		return nil
	}, DefaultE2ETimeout, PollInterval).Should(Succeed())

	return primary
}

func WaitForPrimaryChange(
	namespace string,
	clusterName string,
	oldPrimaryName string,
) corev1.Pod {
	var newPrimary corev1.Pod

	Eventually(func() error {
		primaries, err := GetPrimaryPods(namespace, clusterName)
		if err != nil {
			return err
		}

		if len(primaries) != 1 {
			return fmt.Errorf(
				"expected exactly one primary, got %d",
				len(primaries),
			)
		}

		if primaries[0].Name == oldPrimaryName {
			return fmt.Errorf(
				"primary is still %q",
				oldPrimaryName,
			)
		}

		newPrimary = primaries[0]
		return nil
	}, DefaultE2ETimeout, PollInterval).Should(Succeed())

	return newPrimary
}

func DeletePod(
	namespace string,
	podName string,
) {
	cmd := exec.Command(
		"kubectl",
		"delete",
		"pod",
		podName,
		"--namespace",
		namespace,
		"--wait=false",
	)

	_, err := Run(cmd)
	Expect(err).NotTo(HaveOccurred())
}

func DeleteAllClusterPods(
	namespace string,
	clusterName string,
) {
	pods, err := GetClusterPods(namespace, clusterName)
	Expect(err).NotTo(HaveOccurred())

	for _, pod := range pods {
		DeletePod(namespace, pod.Name)
	}
}

func podReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}

func WaitForPodReady(
	namespace string,
	podName string,
) corev1.Pod {
	var result corev1.Pod

	Eventually(func() error {
		pod, err := GetPod(namespace, podName)
		if err != nil {
			return err
		}

		if !podReady(pod) {
			return fmt.Errorf(
				"pod %q is not Ready",
				podName,
			)
		}

		result = pod
		return nil
	}, DefaultE2ETimeout, PollInterval).Should(Succeed())

	return result
}

// StatefulSet recreates the same pod NAME.
// UID proves that this is actually a new Pod object.
func WaitForPodUIDChange(
	namespace string,
	podName string,
	oldUID types.UID,
) corev1.Pod {
	var recreated corev1.Pod

	Eventually(func() error {
		pod, err := GetPod(namespace, podName)
		if err != nil {
			// NotFound is expected briefly after deletion.
			return err
		}

		if pod.UID == oldUID {
			return fmt.Errorf(
				"pod %q still has old UID %q",
				podName,
				oldUID,
			)
		}

		recreated = pod
		return nil
	}, DefaultE2ETimeout, PollInterval).Should(Succeed())

	return recreated
}

// Use after a standby failure/recovery.
// It verifies that the current primary doesn't unnecessarily move.
func ExpectPrimaryStays(
	namespace string,
	clusterName string,
	expectedPrimary string,
	duration time.Duration,
) {
	Consistently(func() error {
		primaries, err := GetPrimaryPods(namespace, clusterName)
		if err != nil {
			return err
		}

		if len(primaries) != 1 {
			return fmt.Errorf(
				"expected one primary, got %d",
				len(primaries),
			)
		}

		if primaries[0].Name != expectedPrimary {
			return fmt.Errorf(
				"primary changed from %q to %q",
				expectedPrimary,
				primaries[0].Name,
			)
		}

		return nil
	}, duration, PollInterval).Should(Succeed())
}

// Useful while recovering from total outage.
// We don't require a primary at every instant, but split brain is never valid.
func ExpectAtMostOnePrimaryFor(
	namespace string,
	clusterName string,
	duration time.Duration,
) {
	Consistently(func() error {
		primaries, err := GetPrimaryPods(namespace, clusterName)
		if err != nil {
			return err
		}

		if len(primaries) > 1 {
			return fmt.Errorf(
				"split brain: found %d primary pods",
				len(primaries),
			)
		}

		return nil
	}, duration, PollInterval).Should(Succeed())
}

// -----------------------------------------------------------------------------
// Service routing
// -----------------------------------------------------------------------------

func GetOwnedServices(
	namespace string,
	clusterName string,
) ([]corev1.Service, error) {
	var list corev1.ServiceList

	if err := kubectlJSON(
		&list,
		"get",
		"services",
		"--namespace",
		namespace,
		"-o",
		"json",
	); err != nil {
		return nil, err
	}

	services := make([]corev1.Service, 0)

	for _, service := range list.Items {
		for _, owner := range service.OwnerReferences {
			if owner.Kind == "PiHoleCluster" &&
				owner.Name == clusterName {

				services = append(services, service)
				break
			}
		}
	}

	return services, nil
}

func GetServiceEndpointPodNames(
	namespace string,
	serviceName string,
) ([]string, error) {
	var slices discoveryv1.EndpointSliceList

	if err := kubectlJSON(
		&slices,
		"get",
		"endpointslices.discovery.k8s.io",
		"--namespace",
		namespace,
		"--selector",
		fmt.Sprintf(
			"kubernetes.io/service-name=%s",
			serviceName,
		),
		"-o",
		"json",
	); err != nil {
		return nil, err
	}

	names := map[string]struct{}{}

	for _, slice := range slices.Items {
		for _, endpoint := range slice.Endpoints {
			if endpoint.Conditions.Ready != nil &&
				!*endpoint.Conditions.Ready {
				continue
			}

			if endpoint.TargetRef == nil ||
				endpoint.TargetRef.Kind != "Pod" {
				continue
			}

			names[endpoint.TargetRef.Name] = struct{}{}
		}
	}

	result := make([]string, 0, len(names))

	for name := range names {
		result = append(result, name)
	}

	return result, nil
}

// Finds operator-owned Services that explicitly select primary=true
// and verifies that they resolve only to the current primary.
func WaitForPrimaryServicesToTarget(
	namespace string,
	clusterName string,
	primaryName string,
) {
	Eventually(func() error {
		services, err := GetOwnedServices(namespace, clusterName)
		if err != nil {
			return err
		}

		foundPrimaryService := false

		for _, service := range services {
			if service.Spec.Selector[PrimaryLabelKey] != PrimaryLabelValue {
				continue
			}

			foundPrimaryService = true

			pods, err := GetServiceEndpointPodNames(
				namespace,
				service.Name,
			)
			if err != nil {
				return err
			}

			if len(pods) != 1 {
				return fmt.Errorf(
					"service %q has %d ready pod endpoints: %v",
					service.Name,
					len(pods),
					pods,
				)
			}

			if pods[0] != primaryName {
				return fmt.Errorf(
					"service %q targets %q, expected primary %q",
					service.Name,
					pods[0],
					primaryName,
				)
			}
		}

		if !foundPrimaryService {
			return fmt.Errorf(
				"no owned service selects %s=%s",
				PrimaryLabelKey,
				PrimaryLabelValue,
			)
		}

		return nil
	}, DefaultE2ETimeout, PollInterval).Should(Succeed())
}
