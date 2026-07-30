package k8s

import (
	"fmt"
	"log/slog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset kubernetes.Interface
	logger    *slog.Logger
}

func NewClient(logger *slog.Logger) (*Client, error) {
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
	return NewClientFromInterface(cs, logger), nil
}

func NewClientFromInterface(cs kubernetes.Interface, logger *slog.Logger) *Client {
	return &Client{clientset: cs, logger: logger}
}
