package resources

import (
	"context"
	"fmt"
	"slices"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/defaults/conditions"
	"github.com/paldab/pihole-ha-operator/internal/failover"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateFailoverStatus(ctx context.Context, k8sClient client.Client, cluster *v1alpha1.PiHoleCluster, sts *appsv1.StatefulSet) func(*failover.FailoverResult) error {
	return func(failoverResult *failover.FailoverResult) error {
		return UpdateStatusIfChanged(ctx, k8sClient, cluster, sts, failoverResult)
	}
}

func UpdateStatusIfChanged(ctx context.Context, k8sClient client.Client, cluster *v1alpha1.PiHoleCluster, sts *appsv1.StatefulSet, failoverResult *failover.FailoverResult) error {
	desiredStatus := calculateStatus(cluster, sts, failoverResult)

	if !equality.Semantic.DeepEqual(cluster.Status, desiredStatus) {
		originalCluster := cluster.DeepCopy()
		cluster.Status = desiredStatus

		err := k8sClient.Status().Patch(ctx, cluster, client.MergeFrom(originalCluster))
		return err
	}

	return nil
}

// TODO improve this because current status states are ready or progressing
func calculateStatus(cluster *v1alpha1.PiHoleCluster, sts *appsv1.StatefulSet, failoverResult *failover.FailoverResult) v1alpha1.PiHoleClusterStatus {
	desiredReplicas := ptr.Deref(cluster.Spec.Replicas, int32(1))

	newStatus := v1alpha1.PiHoleClusterStatus{
		DesiredReplicas:    desiredReplicas,
		ReadyReplicas:      sts.Status.ReadyReplicas,
		UpdatedReplicas:    sts.Status.UpdatedReplicas,
		CurrentReplicas:    sts.Status.CurrentReplicas,
		ObservedGeneration: cluster.Generation,
		Conditions:         slices.Clone(cluster.Status.Conditions),
	}

	// TODO: handle failover status

	if failoverResult != nil {
		if failoverResult.Leader != nil {
			newStatus.CurrentLeader = &failoverResult.Leader.Name
		} else {
			newStatus.CurrentLeader = nil
		}
	}

	isReady := desiredReplicas == newStatus.ReadyReplicas && desiredReplicas == newStatus.UpdatedReplicas

	if isReady {
		// Ready is true
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:               conditions.TypeReady,
			Status:             metav1.ConditionTrue,
			Reason:             conditions.ReasonClusterReady,
			Message:            conditions.MessageClusterReady,
			ObservedGeneration: cluster.Generation,
		})

		// Progressing is false
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:               conditions.TypeProgressing,
			Status:             metav1.ConditionFalse,
			Reason:             conditions.ReasonClusterReady,
			Message:            conditions.MessageClusterReady,
			ObservedGeneration: cluster.Generation,
		})

		return newStatus
	}

	meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
		Type:               conditions.TypeReady,
		Status:             metav1.ConditionFalse,
		Reason:             conditions.ReasonStatefulSetNotReady,
		ObservedGeneration: cluster.Generation,
		Message: fmt.Sprintf(
			"%d of %d Pihole replicas are ready",
			newStatus.ReadyReplicas,
			desiredReplicas,
		),
	})

	meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
		Type:               conditions.TypeProgressing,
		Status:             metav1.ConditionTrue,
		Reason:             conditions.ReasonStatefulSetNotReady,
		ObservedGeneration: cluster.Generation,
		Message: fmt.Sprintf(
			"%d of %d Pihole replicas are ready",
			newStatus.ReadyReplicas,
			desiredReplicas,
		),
	})

	return newStatus
}
