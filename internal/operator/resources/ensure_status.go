package resources

import (
	"fmt"
	"slices"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults/conditions"
	"github.com/paldab/pihole-ha-operator/internal/operator/failover"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateFailoverStatus(rc *ResourceContext, sts *appsv1.StatefulSet) func(*failover.FailoverResult) error {
	return func(failoverResult *failover.FailoverResult) error {
		return UpdateStatusIfChanged(rc, sts, failoverResult)
	}
}

func UpdateStatusIfChanged(rc *ResourceContext, sts *appsv1.StatefulSet, failoverResult *failover.FailoverResult) error {
	desiredStatus := calculateStatus(rc.Cluster, sts, failoverResult)

	if !equality.Semantic.DeepEqual(rc.Cluster.Status, desiredStatus) {
		// log := logf.FromContext(rc.Ctx)
		// log.Info("patching status",
		// 	"old", rc.Cluster.Status,
		// 	"new", desiredStatus,
		// )
		originalCluster := rc.Cluster.DeepCopy()
		rc.Cluster.Status = desiredStatus

		err := rc.K8sClient.Status().Patch(rc.Ctx, rc.Cluster, client.MergeFrom(originalCluster))
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
		newStatus.FailoverInProgress = &failoverResult.InProgress
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
