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
	CreatedAt time.Time `json:"created_at"`
}

type Container struct {
	Name            string     `json:"name"`
	Image           string     `json:"image"`
	Ready           bool       `json:"ready"`
	Restarts        int32      `json:"restarts"`
	LastExitTime    *time.Time `json:"last_exit_time,omitempty"`
	LastExitReason  *string    `json:"last_exit_reason,omitempty"`
	LastExitMessage *string    `json:"last_exit_message,omitempty"`
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

type EventInvolvedObject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Event struct {
	Type           string              `json:"type"`
	Reason         string              `json:"reason"`
	Message        string              `json:"message"`
	Count          int32               `json:"count"`
	FirstTime      time.Time           `json:"first_time"`
	LastTime       time.Time           `json:"last_time"`
	InvolvedObject EventInvolvedObject `json:"involved_object"`
}
