package k8s

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jdebug14/kube-portal/internal/types"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_getDeploymentStrategy_RollingUpdate_int(t *testing.T) {
	strategy := appsv1.DeploymentStrategy{
		Type: "RollingUpdate",
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(2),
			},
			MaxSurge: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(1),
			},
		},
	}

	results := getDeploymentStrategy(strategy)

	if assert.NotNil(t, results) {
		assert.Equal(t, "RollingUpdate [max unavailable: 2, max surge: 1]", results)
	}
}

func Test_getDeploymentStrategy_RollingUpdate_string(t *testing.T) {
	strategy := appsv1.DeploymentStrategy{
		Type: "RollingUpdate",
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "25%",
			},
			MaxSurge: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "20%",
			},
		},
	}

	results := getDeploymentStrategy(strategy)

	if assert.NotNil(t, results) {
		assert.Equal(t, "RollingUpdate [max unavailable: 25%, max surge: 20%]", results)
	}
}

func Test_getDeploymentStrategy_RollingUpdate_missingField(t *testing.T) {
	strategy := appsv1.DeploymentStrategy{
		Type: "RollingUpdate",
	}

	results := getDeploymentStrategy(strategy)

	if assert.NotNil(t, results) {
		assert.Equal(t, "RollingUpdate", results)
	}
}

func Test_getDeploymentStrategy_RollingUpdate_missingValues(t *testing.T) {
	strategy := appsv1.DeploymentStrategy{
		Type:          "RollingUpdate",
		RollingUpdate: &appsv1.RollingUpdateDeployment{},
	}

	results := getDeploymentStrategy(strategy)

	if assert.NotNil(t, results) {
		assert.Equal(t, "RollingUpdate [max unavailable: <nil>, max surge: <nil>]", results)
	}
}

func Test_getDeploymentStrategy_NotRollingUpdate(t *testing.T) {
	strategy := appsv1.DeploymentStrategy{
		Type: "SomeOtherType",
	}

	results := getDeploymentStrategy(strategy)

	if assert.NotNil(t, results) {
		assert.Equal(t, "SomeOtherType", results)
	}
}

func Test_mapConditions(t *testing.T) {
	now := metav1.Now()
	hourago := metav1.NewTime(now.Time.Add(time.Hour))
	conditions := []appsv1.DeploymentCondition{
		{
			Type:               appsv1.DeploymentAvailable,
			Status:             corev1.ConditionFalse,
			LastUpdateTime:     now,
			LastTransitionTime: hourago,
			Reason:             "MinimumReplicasAvailable",
			Message:            "Deployment has minimum availability",
		},
		{
			Type:               appsv1.DeploymentProgressing,
			Status:             corev1.ConditionTrue,
			LastUpdateTime:     now,
			LastTransitionTime: hourago,
			Reason:             "NewReplicaSetAvailable",
			Message:            "ReplicaSet has successfully progressed",
		},
	}

	results := mapConditions(conditions)

	if assert.NotNil(t, results) {
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "Available", results[0].Type)
		assert.Equal(t, "False", results[0].Status)
		assert.Equal(t, now.Time, results[0].LastUpdateTime)
		assert.Equal(t, hourago.Time, results[0].LastTransitionTime)
		assert.Equal(t, "MinimumReplicasAvailable", results[0].Reason)
		assert.Equal(t, "Deployment has minimum availability", results[0].Message)

		assert.Equal(t, "Progressing", results[1].Type)
		assert.Equal(t, "True", results[1].Status)
		assert.Equal(t, now.Time, results[1].LastUpdateTime)
		assert.Equal(t, hourago.Time, results[1].LastTransitionTime)
		assert.Equal(t, "NewReplicaSetAvailable", results[1].Reason)
		assert.Equal(t, "ReplicaSet has successfully progressed", results[1].Message)
	}
}

func Test_mapConditions_empty(t *testing.T) {
	conditions := []appsv1.DeploymentCondition{}
	results := mapConditions(conditions)
	if assert.NotNil(t, results) {
		assert.Equal(t, 0, len(results))
	}
}

func Test_mapRevisions_happy_path(t *testing.T) {
	now := metav1.Now()
	truth := true
	zero := int32(0)
	two := int32(2)
	three := int32(3)
	four := int32(4)
	five := int32(5)

	replicaset1 := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicaset1",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        "123",
				},
			},
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "2"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: now,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app1",
							Image:   "myrepository/app1",
							Command: []string{"sleep", "100"},
							Args:    []string{"debug"},
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          zero,
			ReadyReplicas:     zero,
			AvailableReplicas: zero,
		},
	}
	replicaset2 := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicaset2",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        "123",
				},
			},
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "3"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: now,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &five,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app1",
							Image:   "myrepository/app1",
							Command: []string{"sleep", "100"},
							Args:    []string{"debug"},
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          four,
			ReadyReplicas:     three,
			AvailableReplicas: two,
		},
	}
	otherReplicaset := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "otherreplicaset",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        "789",
				},
			},
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "1"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &five,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod2",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app1",
							Image: "myrepository/app2",
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          five,
			ReadyReplicas:     five,
			AvailableReplicas: five,
		},
	}

	replicaSetList := appsv1.ReplicaSetList{
		Items: []appsv1.ReplicaSet{*replicaset1, *replicaset2, *otherReplicaset},
	}

	results := mapRevisions(&replicaSetList, "123", slog.New(slog.NewTextHandler(io.Discard, nil)))

	if assert.NotNil(t, results) {
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "replicaset2", results[0].Name)
		assert.Equal(t, int64(3), results[0].Number)
		assert.True(t, results[0].IsCurrent)
		assert.Equal(t, five, results[0].DesiredReplicas)
		assert.Equal(t, four, results[0].CurrentReplicas)
		assert.Equal(t, three, results[0].ReadyReplicas)
		assert.NotNil(t, results[0].PodTemplate)
		assert.Equal(t, now.Time, results[0].CreatedAt)

		assert.Equal(t, "replicaset1", results[1].Name)
		assert.Equal(t, int64(2), results[1].Number)
		assert.False(t, results[1].IsCurrent)
		assert.Equal(t, zero, results[1].DesiredReplicas)
		assert.Equal(t, zero, results[1].CurrentReplicas)
		assert.Equal(t, zero, results[1].ReadyReplicas)
		assert.NotNil(t, results[1].PodTemplate)
		assert.Equal(t, now.Time, results[1].CreatedAt)
	}
}

func Test_mapRevisions_ownerRefNotController(t *testing.T) {
	now := metav1.Now()
	truth := true
	zero := int32(0)
	two := int32(2)
	three := int32(3)
	four := int32(4)
	five := int32(5)

	replicaset1 := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicaset1",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        "123",
				},
			},
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "2"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: now,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app1",
							Image:   "myrepository/app1",
							Command: []string{"sleep", "100"},
							Args:    []string{"debug"},
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          zero,
			ReadyReplicas:     zero,
			AvailableReplicas: zero,
		},
	}
	replicasetBadOwnerRef := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicaset2",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					UID:  "123",
				},
			},
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "3"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: now,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &five,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app1",
							Image:   "myrepository/app1",
							Command: []string{"sleep", "100"},
							Args:    []string{"debug"},
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          four,
			ReadyReplicas:     three,
			AvailableReplicas: two,
		},
	}

	replicaSetList := appsv1.ReplicaSetList{
		Items: []appsv1.ReplicaSet{*replicaset1, *replicasetBadOwnerRef},
	}

	results := mapRevisions(&replicaSetList, "123", slog.New(slog.NewTextHandler(io.Discard, nil)))

	if assert.NotNil(t, results) {
		assert.Equal(t, "replicaset1", results[0].Name)
		assert.Equal(t, int64(2), results[0].Number)
		assert.True(t, results[0].IsCurrent)
		assert.Equal(t, zero, results[0].DesiredReplicas)
		assert.Equal(t, zero, results[0].CurrentReplicas)
		assert.Equal(t, zero, results[0].ReadyReplicas)
		assert.NotNil(t, results[0].PodTemplate)
		assert.Equal(t, now.Time, results[0].CreatedAt)
	}
}

func Test_mapRevisions_missingAnnotationAndReplicas(t *testing.T) {
	now := metav1.Now()
	truth := true
	zero := int32(0)
	two := int32(2)
	three := int32(3)
	four := int32(4)

	replicaset1 := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicaset1",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        "123",
				},
			},
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "2"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: now,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &zero,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app1",
							Image:   "myrepository/app1",
							Command: []string{"sleep", "100"},
							Args:    []string{"debug"},
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          zero,
			ReadyReplicas:     zero,
			AvailableReplicas: zero,
		},
	}
	replicasetBad := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicasetBad",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        "123",
				},
			},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			CreationTimestamp: now,
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "app1",
							Image:   "myrepository/app1",
							Command: []string{"sleep", "100"},
							Args:    []string{"debug"},
						},
					},
				},
			},
		},
		Status: appsv1.ReplicaSetStatus{
			Replicas:          four,
			ReadyReplicas:     three,
			AvailableReplicas: two,
		},
	}

	replicaSetList := appsv1.ReplicaSetList{
		Items: []appsv1.ReplicaSet{*replicaset1, *replicasetBad},
	}

	results := mapRevisions(&replicaSetList, "123", slog.New(slog.NewTextHandler(io.Discard, nil)))

	if assert.NotNil(t, results) {
		assert.Equal(t, 2, len(results))
		assert.Equal(t, "replicasetBad", results[1].Name)
		assert.Equal(t, int64(0), results[1].Number)
		assert.False(t, results[1].IsCurrent)
		assert.Equal(t, zero, results[1].DesiredReplicas)
		assert.Equal(t, four, results[1].CurrentReplicas)
		assert.Equal(t, three, results[1].ReadyReplicas)
		assert.NotNil(t, results[1].PodTemplate)
		assert.Equal(t, now.Time, results[1].CreatedAt)

		assert.Equal(t, "replicaset1", results[0].Name)
		assert.Equal(t, int64(2), results[0].Number)
		assert.True(t, results[0].IsCurrent)
		assert.Equal(t, zero, results[0].DesiredReplicas)
		assert.Equal(t, zero, results[0].CurrentReplicas)
		assert.Equal(t, zero, results[0].ReadyReplicas)
		assert.NotNil(t, results[0].PodTemplate)
		assert.Equal(t, now.Time, results[0].CreatedAt)
	}
}

func Test_mapPodTemplate(t *testing.T) {
	pts := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "pod1",
			Annotations: map[string]string{"my.test/someannotation": "testvalue"},
			Labels:      map[string]string{"hello": "world", "tier": "backend"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "container1",
					Image: "myrepository/app1",
				},
				{
					Name:    "sidecar",
					Image:   "myrepository/sidecar1",
					Command: []string{"sleep", "100"},
					Args:    []string{"debug"},
				},
			},
		},
	}

	results := mapPodTemplate(pts)
	if assert.NotNil(t, results) {
		assert.Equal(t, 1, len(results.Annotations))
		assert.Equal(t, 2, len(results.Labels))
		assert.Equal(t, 2, len(results.Containers))
		assert.Equal(t, "container1", results.Containers[0].Name)
		assert.Equal(t, "myrepository/app1", results.Containers[0].Image)
		assert.Nil(t, results.Containers[0].Command)
		assert.Nil(t, results.Containers[0].Args)
		assert.Equal(t, "sidecar", results.Containers[1].Name)
		assert.Equal(t, "myrepository/sidecar1", results.Containers[1].Image)
		assert.Equal(t, 2, len(results.Containers[1].Command))
		assert.Equal(t, 1, len(results.Containers[1].Args))
	}
}

func Test_mapPodTemplate_emptyContainers(t *testing.T) {
	pts := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "pod1",
			Annotations: map[string]string{"my.test/someannotation": "testvalue"},
			Labels:      map[string]string{"hello": "world", "tier": "backend"},
		},
		Spec: corev1.PodSpec{},
	}

	results := mapPodTemplate(pts)
	if assert.NotNil(t, results) {
		assert.Equal(t, 1, len(results.Annotations))
		assert.Equal(t, 2, len(results.Labels))
		assert.Equal(t, 0, len(results.Containers))
	}
}

func Test_mapPodTemplate_empty(t *testing.T) {
	pts := corev1.PodTemplateSpec{}

	results := mapPodTemplate(pts)
	if assert.NotNil(t, results) {
		assert.Nil(t, results.Annotations)
		assert.Nil(t, results.Labels)
		assert.Equal(t, 0, len(results.Containers))
	}
}

func Test_getRolloutStatus(t *testing.T) {
	now := time.Now()
	tests := []struct {
		caseName   string
		conditions []types.DeploymentCondition
		expected   string
	}{
		{
			caseName: "happy path",
			conditions: []types.DeploymentCondition{
				{
					Type:               "Available",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability",
				},
				{
					Type:               "Progressing",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet has successfully progressed",
				},
			},
			expected: "Running",
		},
		{
			caseName: "ReplicaFailure true always signals Failure",
			conditions: []types.DeploymentCondition{
				{
					Type:               "Available",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability",
				},
				{
					Type:               "Progressing",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet has successfully progressed",
				},
				{
					Type:               "ReplicaFailure",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "SomeFailureReason",
					Message:            "Some failure message",
				},
			},
			expected: "Failure",
		},
		{
			caseName: "no Progressing condition",
			conditions: []types.DeploymentCondition{
				{
					Type:               "Available",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability",
				},
			},
			expected: "Unknown",
		},
		{
			caseName: "Progressing is not True",
			conditions: []types.DeploymentCondition{
				{
					Type:               "Available",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability",
				},
				{
					Type:               "Progressing",
					Status:             "False",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet has successfully progressed",
				},
			},
			expected: "Failure",
		},
		{
			caseName: "Progressing reason does not signal Available",
			conditions: []types.DeploymentCondition{
				{
					Type:               "Available",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability",
				},
				{
					Type:               "Progressing",
					Status:             "True",
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "Some other reason",
					Message:            "Some other message",
				},
			},
			expected: "Pending",
		},
		{caseName: "no conditions", conditions: []types.DeploymentCondition{}, expected: "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.caseName, func(t *testing.T) {
			result := getRolloutStatus(tc.conditions)
			assert.Equal(t, tc.expected, result)
		})
	}

}
