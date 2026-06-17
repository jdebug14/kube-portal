package k8s

import (
	"context"
	"fmt"

	"github.com/jdebug14/kube-portal/internal/types"
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
