package types

import (
	"time"
)

type Namespace struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Deployment struct {
	Name            string    `json:"name"`
	Namespace       string    `json:"namespace"`
	DesiredReplicas int32     `json:"desired_replicas"`
	ReadyReplicas   int32     `json:"ready_replicas"`
	CreatedAt       time.Time `json:"created_at"`
}

type Pod struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Phase     string    `json:"phase"`
	HostNode  string    `json:"host_node"`
	CreatedAt time.Time `json:"created_at"`
}

type Container struct {
	Name           string     `json:"name"`
	Image          string     `json:"image"`
	Ready          bool       `json:"ready"`
	Restarts       int32      `json:"restarts"`
	LastExitCode   int32      `json:"last_exit_code"`
	LastExitReason string     `json:"last_exit_reason"`
	LastFinishedAt *time.Time `json:"last_finished_at"`
}

type PodDetail struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Phase       string            `json:"phase"`
	HostNode    string            `json:"host_node"`
	CreatedAt   time.Time         `json:"created_at"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Containers  []Container       `json:"containers"`
}
