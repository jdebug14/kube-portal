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
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	DesiredReplicas   int32     `json:"desired_replicas"`
	ReadyReplicas     int32     `json:"ready_replicas"`
	AvailableReplicas int32     `json:"available_replicas"`
	CreatedAt         time.Time `json:"created_at"`
}

type DeploymentDetails struct {
	Name              string                `json:"name"`
	Namespace         string                `json:"namespace"`
	Strategy          string                `json:"strategy"`
	DesiredReplicas   int32                 `json:"desired_replicas"`
	UpdatedReplicas   int32                 `json:"updated_replicas"`
	ReadyReplicas     int32                 `json:"ready_replicas"`
	AvailableReplicas int32                 `json:"available_replicas"`
	RolloutStatus     string                `json:"rollout_status"`
	Conditions        []DeploymentCondition `json:"conditions"`
	Revisions         []DeploymentRevision  `json:"revisions"`
	CreatedAt         time.Time             `json:"created_at"`
	Annotations       map[string]string     `json:"annotations"`
	Labels            map[string]string     `json:"labels"`
}

type DeploymentCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	LastUpdateTime     time.Time `json:"last_update_time"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

type DeploymentRevision struct {
	Name            string      `json:"name"`
	Number          int64       `json:"number"`
	IsCurrent       bool        `json:"is_current"`
	DesiredReplicas int32       `json:"desired_replicas"`
	CurrentReplicas int32       `json:"current_replicas"`
	ReadyReplicas   int32       `json:"ready_replicas"`
	PodTemplate     PodTemplate `json:"pod_template"`
	CreatedAt       time.Time   `json:"created_at"`
}

type PodTemplate struct {
	Annotations map[string]string   `json:"annotations,omitempty"`
	Labels      map[string]string   `json:"labels,omitempty"`
	Containers  []ContainerTemplate `json:"containers"`
}

type ContainerTemplate struct {
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

type Pod struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Phase     string    `json:"phase"`
	CreatedAt time.Time `json:"created_at"`
}

type PodDetails struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Phase       string            `json:"phase"`
	HostNode    string            `json:"host_node"`
	CreatedAt   time.Time         `json:"created_at"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
	Containers  []Container       `json:"containers"`
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
