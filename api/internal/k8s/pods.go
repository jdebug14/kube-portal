package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/jdebug14/kube-portal/internal/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) ListPods(ctx context.Context, namespace string) ([]types.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	results := make([]types.Pod, 0, len(podList.Items))
	for _, p := range podList.Items {
		results = append(results, types.Pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Phase:     string(p.Status.Phase),
			CreatedAt: p.CreationTimestamp.Time,
		})
	}
	return results, nil
}

func (c *Client) GetPodDetails(ctx context.Context, namespace string, podName string) (types.PodDetail, error) {
	podDetails, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return types.PodDetail{}, fmt.Errorf("failed to get pod details: %w", err)
	}
	result := types.PodDetail{
		Name:        podDetails.Name,
		Namespace:   podDetails.Namespace,
		Phase:       string(podDetails.Status.Phase),
		HostNode:    podDetails.Spec.NodeName,
		CreatedAt:   podDetails.CreationTimestamp.Time,
		Annotations: podDetails.Annotations,
		Labels:      podDetails.Labels,
		Containers:  mapContainers(podDetails.Spec.Containers, podDetails.Status.ContainerStatuses),
	}
	return result, nil
}

func mapContainers(containers []v1.Container, statuses []v1.ContainerStatus) []types.Container {
	statusMap := make(map[string]v1.ContainerStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}
	results := make([]types.Container, 0, len(containers))
	for _, c := range containers {
		status := statusMap[c.Name]
		var lastExitTime *time.Time
		var lastExitReason *string
		var lastExitMessage *string

		if status.LastTerminationState.Terminated != nil {
			lastExitTimeValue := status.LastTerminationState.Terminated.FinishedAt.Time
			lastExitReasonValue := status.LastTerminationState.Terminated.Reason
			lastExitMessageValue := status.LastTerminationState.Terminated.Message
			lastExitTime = &lastExitTimeValue
			lastExitReason = &lastExitReasonValue
			lastExitMessage = &lastExitMessageValue

		}
		results = append(results, types.Container{
			Name:            c.Name,
			Image:           c.Image,
			Ready:           status.Ready,
			Restarts:        status.RestartCount,
			LastExitTime:    lastExitTime,
			LastExitReason:  lastExitReason,
			LastExitMessage: lastExitMessage,
		})
	}
	return results
}
