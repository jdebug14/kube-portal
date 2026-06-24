package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jdebug14/kube-portal/internal/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset kubernetes.Interface
}

func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			&clientcmd.ConfigOverrides{},
		)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &Client{clientset: cs}, nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]types.Namespace, error) {
	namespaceList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch namespaces: %w", err)
	}
	results := make([]types.Namespace, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		results = append(results, types.Namespace{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			CreatedAt: ns.CreationTimestamp.Time,
		})
	}
	return results, nil
}

func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]types.Deployment, error) {
	deploymentList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployments: %w", err)
	}
	results := make([]types.Deployment, 0, len(deploymentList.Items))
	for _, d := range deploymentList.Items {
		var desiredReplicas int32
		if d.Spec.Replicas != nil {
			desiredReplicas = *d.Spec.Replicas
		}
		results = append(results, types.Deployment{
			Name:            d.Name,
			Namespace:       d.Namespace,
			DesiredReplicas: desiredReplicas,
			ReadyReplicas:   d.Status.ReadyReplicas,
			CreatedAt:       d.CreationTimestamp.Time,
		})
	}
	return results, nil
}

func (c *Client) ListPods(ctx context.Context, namespace string) ([]types.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %w", err)
	}
	results := make([]types.Pod, 0, len(podList.Items))
	for _, p := range podList.Items {
		results = append(results, types.Pod{
			Name:      p.Name,
			Namespace: p.Namespace,
			Phase:     string(p.Status.Phase),
			HostNode:  p.Spec.NodeName,
			CreatedAt: p.CreationTimestamp.Time,
		})
	}
	return results, nil
}

func (c *Client) GetPodDetail(ctx context.Context, namespace string, podName string) (types.PodDetail, error) {
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

func (c *Client) ListNamespaceEvents(ctx context.Context, namespace string, involvedObjectName string) ([]types.Event, error) {
	var listOpts metav1.ListOptions
	if involvedObjectName != "" {
		listOpts = metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + involvedObjectName,
		}
	} else {
		listOpts = metav1.ListOptions{}
	}
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch namespace events: %w", err)
	}
	results := make([]types.Event, 0, len(events.Items))
	for _, e := range events.Items {
		results = append(results, types.Event{
			Type:      e.Type,
			Reason:    e.Reason,
			Message:   e.Message,
			Count:     e.Count,
			FirstTime: e.FirstTimestamp.Time,
			LastTime:  e.LastTimestamp.Time,
			InvolvedObject: types.EventInvolvedObject{
				Kind:      e.InvolvedObject.Kind,
				Name:      e.InvolvedObject.Name,
				Namespace: e.InvolvedObject.Namespace,
			},
		})
	}
	sort.Slice(results, func(a, b int) bool {
		return results[a].LastTime.After(results[b].LastTime)
	})
	return results, nil
}

func mapContainers(containers []v1.Container, statuses []v1.ContainerStatus) []types.Container {
	statusMap := make(map[string]v1.ContainerStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}
	results := make([]types.Container, 0, len(containers))
	for _, c := range containers {
		status := statusMap[c.Name]
		var lastExitCode int32
		var lastExitReason string
		var lastFinishedAt *time.Time

		if status.LastTerminationState.Terminated != nil {
			lastExitCode = status.LastTerminationState.Terminated.ExitCode
			lastExitReason = status.LastTerminationState.Terminated.Reason
			lastFinishedAt = &status.LastTerminationState.Terminated.FinishedAt.Time
		}
		results = append(results, types.Container{
			Name:           c.Name,
			Image:          c.Image,
			Ready:          status.Ready,
			Restarts:       status.RestartCount,
			LastExitCode:   lastExitCode,
			LastExitReason: lastExitReason,
			LastFinishedAt: lastFinishedAt,
		})
	}
	return results
}
