package k8s

import (
	"context"
	"fmt"
	"sort"

	"github.com/jdebug14/kube-portal/internal/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) ListEvents(ctx context.Context, namespace string, involvedObjectName string) ([]types.Event, error) {
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
		return nil, fmt.Errorf("failed to list events: %w", err)
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
