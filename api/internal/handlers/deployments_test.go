package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jdebug14/kube-portal/internal/types"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8stesting "k8s.io/client-go/testing"
)

func TestListDeployments_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
	replicas := int32(5)
	coredns := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "kube-system",
			Name:              "coredns",
			CreationTimestamp: now,
		},
	}
	app1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "app1",
			CreationTimestamp: now,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 3,
		},
	}
	h, _ := setupHandler(t, coredns, app1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListDeployments(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var deployments []types.Deployment
	err = json.Unmarshal(body, &deployments)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deployments))
	assert.Equal(t, "app1", deployments[0].Name)
	assert.True(t, now.Time.Equal(deployments[0].CreatedAt))
	assert.Equal(t, replicas, deployments[0].DesiredReplicas)
	assert.Equal(t, int32(3), deployments[0].AvailableReplicas)
}

func TestListDeployments_NilReplicas(t *testing.T) {
	// arrange
	now := metav1.Now()
	coredns := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "kube-system",
			Name:              "coredns",
			CreationTimestamp: now,
		},
	}
	app1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "app1",
			CreationTimestamp: now,
		},
		Spec: appsv1.DeploymentSpec{},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 3,
		},
	}
	h, _ := setupHandler(t, coredns, app1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListDeployments(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var deployments []types.Deployment
	err = json.Unmarshal(body, &deployments)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deployments))
	assert.Equal(t, "app1", deployments[0].Name)
	assert.True(t, now.Time.Equal(deployments[0].CreatedAt))
	assert.Equal(t, int32(0), deployments[0].DesiredReplicas)
	assert.Equal(t, int32(3), deployments[0].AvailableReplicas)
}

func TestListDeployments_None(t *testing.T) {
	// arrange
	app1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "othernamespace",
			Name:              "app1",
			CreationTimestamp: metav1.Now(),
		},
	}
	h, _ := setupHandler(t, app1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListDeployments(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var deployments []types.Deployment
	err = json.Unmarshal(body, &deployments)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(deployments))
}

func TestListDeployments_BadRequest(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "Invalidnamespace"})
	w := httptest.NewRecorder()

	// act
	h.ListDeployments(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Contains(t, errorResponse.Message, "invalid namespace")
}

func TestListDeployments_Error(t *testing.T) {
	// arrange
	h, fakeCS := setupHandler(t)
	fakeCS.PrependReactor("list", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListDeployments(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	assert.Equal(t, "failed to retrieve deployments", errorResponse.Message)
}

func TestGetDeploymentDetails_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
	five := int32(5)
	four := int32(4)
	three := int32(3)
	two := int32(2)
	one := int32(1)
	zero := int32(0)
	iors := intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: two,
	}
	truth := true

	deployment1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "deployment1",
			CreationTimestamp: now,
			Annotations:       map[string]string{"deployment.kubernetes.io/revision": "1"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
			UID:               "123",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &five,
			Strategy: appsv1.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &iors,
					MaxSurge:       &iors,
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          five,
			UpdatedReplicas:   four,
			ReadyReplicas:     three,
			AvailableReplicas: two,
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:               appsv1.DeploymentAvailable,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "MinimumReplicasAvailable",
					Message:            "Deployment has minimum availability",
				},
				{
					Type:               appsv1.DeploymentProgressing,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     now,
					LastTransitionTime: now,
					Reason:             "NewReplicaSetAvailable",
					Message:            "ReplicaSet has successfully progressed",
				},
			},
		},
	}
	replicaset1 := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "replicaset1",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       "Deployment",
					Controller: &truth,
					UID:        deployment1.UID,
				},
			},
			Annotations: map[string]string{"deployment.kubernetes.io/revision": "1"},
			Labels:      map[string]string{"hello": "world", "tier": "backend"},
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
			Replicas:          one,
			ReadyReplicas:     one,
			AvailableReplicas: one,
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
					UID:        deployment1.UID,
				},
			},
			Annotations: map[string]string{"deployment.kubernetes.io/revision": "2"},
			Labels:      map[string]string{"hello": "world", "tier": "backend"},
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
			Replicas:          five,
			ReadyReplicas:     five,
			AvailableReplicas: five,
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
			Annotations: map[string]string{"deployment.kubernetes.io/revision": "1"},
			Labels:      map[string]string{"hello": "world", "tier": "backend"},
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
	h, _ := setupHandler(t, deployment1, replicaset1, replicaset2, otherReplicaset)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "deployment1"})
	w := httptest.NewRecorder()

	// act
	h.GetDeploymentDetails(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		assert.NoError(t, err)
		var deploymentDetails types.DeploymentDetails
		err = json.Unmarshal(body, &deploymentDetails)

		assert.NoError(t, err)
		assert.Equal(t, "deployment1", deploymentDetails.Name)
		assert.True(t, now.Time.Equal(deploymentDetails.CreatedAt))
		assert.Equal(t, 1, len(deploymentDetails.Annotations))
		assert.Equal(t, 2, len(deploymentDetails.Labels))
		assert.Equal(t, "RollingUpdate [max unavailable: 2, max surge: 2]", deploymentDetails.Strategy)
		assert.Equal(t, five, deploymentDetails.DesiredReplicas)
		assert.Equal(t, four, deploymentDetails.UpdatedReplicas)
		assert.Equal(t, three, deploymentDetails.ReadyReplicas)
		assert.Equal(t, two, deploymentDetails.AvailableReplicas)
		assert.Equal(t, 2, len(deploymentDetails.Conditions))
		assert.Equal(t, "Running", deploymentDetails.RolloutStatus)
		assert.Equal(t, 2, len(deploymentDetails.Revisions))
		assert.Equal(t, "replicaset2", deploymentDetails.Revisions[0].Name)
		assert.True(t, deploymentDetails.Revisions[0].IsCurrent)
		assert.Equal(t, five, deploymentDetails.Revisions[0].DesiredReplicas)
		assert.Equal(t, five, deploymentDetails.Revisions[0].CurrentReplicas)
		assert.Equal(t, five, deploymentDetails.Revisions[0].ReadyReplicas)
		assert.Equal(t, 1, len(deploymentDetails.Revisions[0].PodTemplate.Containers))
		assert.Equal(t, "replicaset1", deploymentDetails.Revisions[1].Name)
		assert.False(t, deploymentDetails.Revisions[1].IsCurrent)
		assert.Equal(t, zero, deploymentDetails.Revisions[1].DesiredReplicas)
		assert.Equal(t, one, deploymentDetails.Revisions[1].CurrentReplicas)
		assert.Equal(t, one, deploymentDetails.Revisions[1].ReadyReplicas)
		assert.Equal(t, 1, len(deploymentDetails.Revisions[1].PodTemplate.Containers))
	}
}

func TestGetDeploymentDetails_Error_getDeployments(t *testing.T) {
	// arrange
	h, fakeCS := setupHandler(t)
	fakeCS.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "app1-87745fd6c1-9gd3s"})
	w := httptest.NewRecorder()

	// act
	h.GetDeploymentDetails(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	assert.Equal(t, "failed to retrieve deployment details", errorResponse.Message)
}

func TestGetDeploymentDetails_Error_getReplicaSets(t *testing.T) {
	// arrange
	deployment1 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "deployment1",
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 3,
		},
	}
	h, fakeCS := setupHandler(t, deployment1)
	fakeCS.PrependReactor("list", "replicasets", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "deployment1"})
	w := httptest.NewRecorder()

	// act
	h.GetDeploymentDetails(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	assert.Equal(t, "failed to retrieve deployment details", errorResponse.Message)
}

func TestGetDeploymentDetails_BadRequest_namespace(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "Invalidnamespace", "name": "deployment1"})
	w := httptest.NewRecorder()

	// act
	h.GetDeploymentDetails(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Contains(t, errorResponse.Message, "invalid namespace")
}

func TestGetDeploymentDetails_DNE(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "deployment1"})
	w := httptest.NewRecorder()

	// act
	h.GetDeploymentDetails(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, errorResponse.Code)
	assert.Contains(t, errorResponse.Message, "deployment not found")
}

func TestGetDeploymentDetails_BadRequest_name(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "&%$"})
	w := httptest.NewRecorder()

	// act
	h.GetDeploymentDetails(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var errorResponse errorResponse
	err = json.Unmarshal(body, &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Contains(t, errorResponse.Message, "invalid deployment")
}
