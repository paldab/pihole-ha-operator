package status

import (
	"fmt"
	"slices"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/failover"
	"github.com/paldab/pihole-ha-operator/internal/operator/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateClusterStatus(rc *resources.ResourceContext, sts *appsv1.StatefulSet, failoverResult *failover.FailoverResult) error {
	desiredStatus := calculateStatus(rc.Cluster, sts, failoverResult)

	if !equality.Semantic.DeepEqual(rc.Cluster.Status, desiredStatus) {
		originalCluster := rc.Cluster.DeepCopy()
		rc.Cluster.Status = desiredStatus

		err := rc.K8sClient.Status().Patch(rc.Ctx, rc.Cluster, client.MergeFrom(originalCluster))
		return err
	}

	return nil
}

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

	newStatus.Statistics = updateStatisticsStatus(cluster)

	failingOver := false
	if failoverResult != nil {
		failingOver = failoverResult.InProgress

		if failoverResult.Leader != nil {
			newStatus.CurrentLeader = &failoverResult.Leader.Name
		} else {
			newStatus.CurrentLeader = nil
		}
	}

	if !failingOver {
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:               TypeFailingOver,
			Status:             metav1.ConditionFalse,
			Reason:             ReasonFailoverCompleted,
			Message:            "Pihole primary failover is complete",
			ObservedGeneration: cluster.Generation,
		})
	}

	switch {
	// adding this case just because otherwise it get's ready when failing over
	case failingOver:
		setClusterNotReady(
			&newStatus,
			cluster.Generation,
			ReasonFailoverInProgress,
			"Pihole primary failover is in progress",
		)

	case newStatus.UpdatedReplicas != desiredReplicas:
		setClusterNotReady(
			&newStatus,
			cluster.Generation,
			ReasonStatefulSetUpdating,
			fmt.Sprintf(
				"%d of %d Pihole replicas are updated",
				newStatus.UpdatedReplicas,
				desiredReplicas,
			),
		)

	case newStatus.ReadyReplicas != desiredReplicas:
		setClusterNotReady(
			&newStatus,
			cluster.Generation,
			ReasonStatefulSetNotReady,
			fmt.Sprintf(
				"%d of %d Pihole replicas are ready",
				newStatus.ReadyReplicas,
				desiredReplicas,
			),
		)

	case newStatus.CurrentLeader == nil:
		setClusterNotReady(
			&newStatus,
			cluster.Generation,
			failover.ReasonLeaderUnavailable,
			"No Pihole primary is currently available",
		)

	default:
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:               ConditionClusterReady,
			Status:             metav1.ConditionTrue,
			Reason:             string(ReasonClusterHealthy),
			Message:            "Pihole cluster is ready",
			ObservedGeneration: cluster.Generation,
		})

	}

	return newStatus
}

func setClusterNotReady(status *v1alpha1.PiHoleClusterStatus, observedGeneration int64, reason, message string) {
	meta.SetStatusCondition(&status.Conditions, metav1.Condition{
		Type:               TypeReady,
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		ObservedGeneration: observedGeneration,
		Message:            message,
	})
}

func updateStatisticsStatus(cluster *v1alpha1.PiHoleCluster) v1alpha1.StatisticsStatus {
	var statisticsStatus = v1alpha1.StatisticsStatus{
		Mode: defaults.StatisticsMode,
	}

	clusterStats := cluster.Spec.Statistics

	if clusterStats == nil {
		return statisticsStatus
	}

	statisticsStatus.Mode = clusterStats.Mode
	if clusterStats.Mode != v1alpha1.StatsModeExternal {
		return statisticsStatus
	}

	statisticsStatus.External = &v1alpha1.ExternalStatisticsStatus{
		BatchSize:       defaults.StatisticsExportBatchSize,
		IntervalSeconds: defaults.StatisticsExportInterval,
	}

	if clusterStats.External == nil {
		return statisticsStatus
	}

	if clusterStats.External.BatchSize > 0 {
		statisticsStatus.External.BatchSize = clusterStats.External.BatchSize
	}

	if clusterStats.External.IntervalSeconds >= 60 {
		statisticsStatus.External.IntervalSeconds = clusterStats.External.IntervalSeconds
	}

	return statisticsStatus
}
