package status

import (
	"context"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetConfigReadyCondition(
	ctx context.Context,
	k8sClient client.Client,
	config *v1alpha1.PiHoleConfig,
	status metav1.ConditionStatus,
	reason string,
	message string,
) error {
	originalConfig := config.DeepCopy()

	meta.SetStatusCondition(&config.Status.Conditions, metav1.Condition{
		Type:               ConditionConfigReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: originalConfig.Generation,
	})

	config.Status.ObservedGeneration = config.Generation

	if equality.Semantic.DeepEqual(config.Status, originalConfig.Status) {
		return nil
	}

	return k8sClient.Status().Patch(
		ctx,
		config,
		client.MergeFrom(originalConfig),
	)
}
