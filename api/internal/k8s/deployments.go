package k8s

import (
	"context"
	"fmt"

	"github.com/jdebug14/kube-portal/internal/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Name:            d.Name,
			Namespace:       d.Namespace,
			DesiredReplicas: desiredReplicas,
			ReadyReplicas:   d.Status.ReadyReplicas,
			CreatedAt:       d.CreationTimestamp.Time,
		})
	}
	return results, nil
}
