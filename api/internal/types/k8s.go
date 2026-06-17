package types

import "time"

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
