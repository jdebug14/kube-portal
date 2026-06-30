package k8s

import (
	"context"
	"fmt"

	"github.com/jdebug14/kube-portal/internal/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) ListNamespaces(ctx context.Context) ([]types.Namespace, error) {
	namespaceList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
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
