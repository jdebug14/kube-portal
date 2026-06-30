package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
)

func (c *Client) GetPodLogs(ctx context.Context, namespace string, podName string, container string, tailLines int64) (string, error) {
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer stream.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, stream)
	if err != nil {
		return "", fmt.Errorf("failed to parse logs stream: %w", err)
	}
	return buf.String(), nil
}
