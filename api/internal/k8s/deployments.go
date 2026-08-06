package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"

	"github.com/jdebug14/kube-portal/internal/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]types.Deployment, error) {
	deploymentList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	results := make([]types.Deployment, 0, len(deploymentList.Items))
	for _, d := range deploymentList.Items {
		// it should never happen that this is returned as nil as there is a default of 1,
		var desiredReplicas int32
		if d.Spec.Replicas != nil {
			desiredReplicas = *d.Spec.Replicas
		}
		results = append(results, types.Deployment{
			Name:              d.Name,
			Namespace:         d.Namespace,
			DesiredReplicas:   desiredReplicas,
			ReadyReplicas:     d.Status.ReadyReplicas,
			AvailableReplicas: d.Status.AvailableReplicas,
			CreatedAt:         d.CreationTimestamp.Time,
		})
	}
	return results, nil
}

func (c *Client) GetDeploymentDetails(ctx context.Context, namespace string, deploymentName string) (types.DeploymentDetails, error) {
	d, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return types.DeploymentDetails{}, fmt.Errorf("failed to get deployment %q: %w", deploymentName, err)
	}
	replicasetList, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return types.DeploymentDetails{}, fmt.Errorf("failed to list replicasets for namespace %q: %w", namespace, err)
	}

	var desiredReplicas int32
	if d.Spec.Replicas != nil {
		desiredReplicas = *d.Spec.Replicas
	}
	deploymentConditions := mapConditions(d.Status.Conditions)
	result := types.DeploymentDetails{
		Name:              d.Name,
		Namespace:         d.Namespace,
		Strategy:          getDeploymentStrategy(d.Spec.Strategy),
		DesiredReplicas:   desiredReplicas,
		UpdatedReplicas:   d.Status.UpdatedReplicas,
		ReadyReplicas:     d.Status.ReadyReplicas,
		AvailableReplicas: d.Status.AvailableReplicas,
		Conditions:        deploymentConditions,
		Revisions:         mapRevisions(replicasetList, d.ObjectMeta.UID, c.logger),
		RolloutStatus:     getRolloutStatus(deploymentConditions),
		CreatedAt:         d.CreationTimestamp.Time,
		Annotations:       d.Annotations,
		Labels:            d.Labels,
	}
	return result, nil
}

func getDeploymentStrategy(strategy appsv1.DeploymentStrategy) string {
	strategyType := string(strategy.Type)
	rollingUpdate := strategy.RollingUpdate
	if rollingUpdate != nil {
		return fmt.Sprintf(strategyType+" [max unavailable: %s, max surge: %s]", rollingUpdate.MaxUnavailable.String(), rollingUpdate.MaxSurge.String())
	}
	return string(strategyType)
}

func mapConditions(deploymentConditions []appsv1.DeploymentCondition) []types.DeploymentCondition {
	results := make([]types.DeploymentCondition, 0, len(deploymentConditions))
	for _, c := range deploymentConditions {
		results = append(results, types.DeploymentCondition{
			Type:               string(c.Type),
			Status:             string(c.Status),
			Reason:             c.Reason,
			Message:            c.Message,
			LastUpdateTime:     c.LastUpdateTime.Time,
			LastTransitionTime: c.LastTransitionTime.Time,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].LastUpdateTime.After(results[j].LastUpdateTime)
	})
	return results
}

func mapRevisions(replicaSets *appsv1.ReplicaSetList, deploymentUID k8stypes.UID, logger *slog.Logger) []types.DeploymentRevision {
	results := make([]types.DeploymentRevision, 0, len(replicaSets.Items))
	for _, rs := range replicaSets.Items {
		owner := false
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.Controller != nil && *ownerRef.Controller && ownerRef.UID == deploymentUID {
				owner = true
				break
			}
		}
		if !owner {
			continue
		}

		revisionNumber, ok := getRevisionNumber(rs.Annotations)
		if !ok {
			logger.Warn("replicaset missing revision annotation", "replicaset", rs.Name)
		}

		var desiredReplicas int32
		if rs.Spec.Replicas != nil {
			desiredReplicas = *rs.Spec.Replicas
		}

		results = append(results, types.DeploymentRevision{
			Name:            rs.Name,
			Number:          revisionNumber,
			DesiredReplicas: desiredReplicas,
			CurrentReplicas: rs.Status.Replicas,
			ReadyReplicas:   rs.Status.ReadyReplicas,
			PodTemplate:     mapPodTemplate(rs.Spec.Template),
			CreatedAt:       rs.CreationTimestamp.Time,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Number > results[j].Number
	})
	if len(results) > 0 {
		results[0].IsCurrent = true
	}

	return results
}

func getRevisionNumber(annotations map[string]string) (int64, bool) {
	raw, ok := annotations["deployment.kubernetes.io/revision"]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func mapPodTemplate(pts corev1.PodTemplateSpec) types.PodTemplate {
	containers := make([]types.ContainerTemplate, 0, len(pts.Spec.Containers))
	for _, c := range pts.Spec.Containers {
		containers = append(containers, types.ContainerTemplate{
			Name:    c.Name,
			Image:   c.Image,
			Command: c.Command,
			Args:    c.Args,
		})
	}
	result := types.PodTemplate{
		Annotations: pts.Annotations,
		Labels:      pts.Labels,
		Containers:  containers,
	}
	return result
}

func getRolloutStatus(conditions []types.DeploymentCondition) string {
	if len(conditions) == 0 {
		return "Unknown"
	}

	var progressing *types.DeploymentCondition
	for i, dc := range conditions {
		if dc.Type == "ReplicaFailure" {
			return "Failure"
		}
		if dc.Type == "Progressing" {
			progressing = &conditions[i]
		}
	}

	if progressing == nil {
		return "Unknown"
	}
	if progressing.Status != "True" {
		return "Failure"
	}
	if progressing.Reason == "NewReplicaSetAvailable" {
		return "Running"
	}
	return "Pending"
}
