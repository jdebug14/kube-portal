package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jdebug14/kube-portal/internal/k8s"
	"github.com/jdebug14/kube-portal/internal/types"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func setupHandler(t *testing.T, objects ...runtime.Object) (*Handler, *fake.Clientset) {
	t.Helper()
	fakeCS := fake.NewSimpleClientset(objects...)
	k8sClient := k8s.NewClientFromInterface(fakeCS)
	h := NewHandler(k8sClient, slog.Default())
	return h, fakeCS
}

func newRequestWithParams(t *testing.T, method, url string, rParams map[string]string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, url, nil)
	rctx := chi.NewRouteContext()
	for k, v := range rParams {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestListNamespaces_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
	defaultNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "default",
			CreationTimestamp: now,
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}
	terminatingNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "helloworld",
			CreationTimestamp: now,
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceTerminating,
		},
	}
	h, _ := setupHandler(t, defaultNamespace, terminatingNamespace)
	r := httptest.NewRequest(http.MethodGet, "/some/test/request", nil)
	w := httptest.NewRecorder()

	// act
	h.ListNamespaces(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var namespaces []types.Namespace
	err = json.Unmarshal(body, &namespaces)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(namespaces))
	assert.Equal(t, "default", namespaces[0].Name)
	assert.Equal(t, "Active", namespaces[0].Status)
	assert.True(t, now.Time.Equal(namespaces[0].CreatedAt))
	assert.Equal(t, "helloworld", namespaces[1].Name)
	assert.Equal(t, "Terminating", namespaces[1].Status)
	assert.True(t, now.Time.Equal(namespaces[1].CreatedAt))
}

func TestListNamespaces_Empty(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/some/test/request", nil)
	w := httptest.NewRecorder()

	// act
	h.ListNamespaces(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var namespaces []types.Namespace
	err = json.Unmarshal(body, &namespaces)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(namespaces))
}

func TestListNamespaces_Error(t *testing.T) {
	// arrange
	h, fakeCS := setupHandler(t)
	fakeCS.PrependReactor("list", "namespaces", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := httptest.NewRequest(http.MethodGet, "/some/test/request", nil)
	w := httptest.NewRecorder()

	// act
	h.ListNamespaces(w, r)

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
	assert.Equal(t, "failed to fetch namespaces", errorResponse.Message)
}

func TestListDeployments_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
	replicasValue := int32(5)
	replicas := &replicasValue
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
			Replicas: replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 3,
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
	assert.Equal(t, replicasValue, deployments[0].DesiredReplicas)
	assert.Equal(t, int32(3), deployments[0].ReadyReplicas)
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
	assert.Equal(t, "failed to fetch deployments", errorResponse.Message)
}

func TestListPods_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
	coredns := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "kube-system",
			Name:              "coredns",
			CreationTimestamp: now,
		},
	}
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "app1",
			CreationTimestamp: now,
		},
		Spec: corev1.PodSpec{
			NodeName: "worker-1",
		},
		Status: corev1.PodStatus{
			Phase: "Running",
		},
	}
	h, _ := setupHandler(t, coredns, pod1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListPods(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var pods []types.Pod
	err = json.Unmarshal(body, &pods)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pods))
	assert.Equal(t, "app1", pods[0].Name)
	assert.True(t, now.Time.Equal(pods[0].CreatedAt))
	assert.Equal(t, "Running", pods[0].Phase)
}

func TestListPods_None(t *testing.T) {
	// arrange
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "othernamespace",
			Name:              "app1",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			NodeName: "worker-1",
		},
		Status: corev1.PodStatus{
			Phase: "Running",
		},
	}
	h, _ := setupHandler(t, pod1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListPods(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var pods []types.Pod
	err = json.Unmarshal(body, &pods)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(pods))
}

func TestListPods_BadRequest(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "Invalidnamespace"})
	w := httptest.NewRecorder()

	// act
	h.ListPods(w, r)

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

func TestListPods_Error(t *testing.T) {
	// arrange
	h, fakeCS := setupHandler(t)
	fakeCS.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListPods(w, r)

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
	assert.Equal(t, "failed to fetch pods", errorResponse.Message)
}

func TestGetPodDetails_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
	container1 := &corev1.Container{
		Name:  "app1",
		Image: "myrepository/app1",
	}
	status1 := &corev1.ContainerStatus{
		Name:         "app1",
		Ready:        true,
		RestartCount: 12,
		LastTerminationState: corev1.ContainerState{
			Terminated: &corev1.ContainerStateTerminated{
				FinishedAt: now,
				Reason:     "Completed",
			},
		},
	}
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "pod1",
			CreationTimestamp: now,
			Annotations:       map[string]string{"some.test/anno.tation": "123"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
		},
		Spec: corev1.PodSpec{
			NodeName:   "worker-1",
			Containers: []corev1.Container{*container1},
		},
		Status: corev1.PodStatus{
			Phase:             "Running",
			ContainerStatuses: []corev1.ContainerStatus{*status1},
		},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "pod2",
			CreationTimestamp: now,
			Annotations:       map[string]string{"some.test/anno.tation": "123"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
		},
		Spec: corev1.PodSpec{
			NodeName:   "worker-2",
			Containers: []corev1.Container{*container1},
		},
		Status: corev1.PodStatus{
			Phase:             "Running",
			ContainerStatuses: []corev1.ContainerStatus{*status1},
		},
	}
	h, _ := setupHandler(t, pod1, pod2)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "pn": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetail(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var podDetails types.PodDetail
	err = json.Unmarshal(body, &podDetails)
	assert.NoError(t, err)
	assert.Equal(t, "pod1", podDetails.Name)
	assert.True(t, now.Time.Equal(podDetails.CreatedAt))
	assert.Equal(t, "Running", podDetails.Phase)
	assert.Equal(t, "worker-1", podDetails.HostNode)
	assert.Equal(t, 1, len(podDetails.Annotations))
	assert.Equal(t, 2, len(podDetails.Labels))
	assert.Equal(t, 1, len(podDetails.Containers))
	assert.Equal(t, "app1", podDetails.Containers[0].Name)
	assert.Equal(t, "myrepository/app1", podDetails.Containers[0].Image)
	assert.True(t, podDetails.Containers[0].Ready)
	assert.Equal(t, int32(12), podDetails.Containers[0].Restarts)
	assert.True(t, now.Time.Equal(*podDetails.Containers[0].LastExitTime))
	assert.Equal(t, "Completed", *podDetails.Containers[0].LastExitReason)
}

func TestGetPodDetails_NeverTerminated(t *testing.T) {
	// arrange
	now := metav1.Now()
	container1 := &corev1.Container{
		Name:  "app1",
		Image: "myrepository/app1",
	}
	status1 := &corev1.ContainerStatus{
		Name:         "app1",
		Ready:        true,
		RestartCount: 0,
	}
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "pod1",
			CreationTimestamp: now,
			Annotations:       map[string]string{"some.test/anno.tation": "123"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
		},
		Spec: corev1.PodSpec{
			NodeName:   "worker-2",
			Containers: []corev1.Container{*container1},
		},
		Status: corev1.PodStatus{
			Phase:             "Running",
			ContainerStatuses: []corev1.ContainerStatus{*status1},
		},
	}
	h, _ := setupHandler(t, pod1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "pn": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetail(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var podDetails types.PodDetail
	err = json.Unmarshal(body, &podDetails)
	assert.NoError(t, err)
	assert.Equal(t, "pod1", podDetails.Name)
	assert.True(t, now.Time.Equal(podDetails.CreatedAt))
	assert.Equal(t, "Running", podDetails.Phase)
	assert.Equal(t, "worker-2", podDetails.HostNode)
	assert.Equal(t, 1, len(podDetails.Annotations))
	assert.Equal(t, 2, len(podDetails.Labels))
	assert.Equal(t, 1, len(podDetails.Containers))
	assert.Equal(t, "app1", podDetails.Containers[0].Name)
	assert.Equal(t, "myrepository/app1", podDetails.Containers[0].Image)
	assert.True(t, podDetails.Containers[0].Ready)
	assert.Equal(t, int32(0), podDetails.Containers[0].Restarts)
	assert.Nil(t, podDetails.Containers[0].LastExitTime)
	assert.Nil(t, podDetails.Containers[0].LastExitReason)
}

func TestGetPodDetails_BadRequest_NS(t *testing.T) {
	// arrange
	container1 := &corev1.Container{
		Name:  "app1",
		Image: "myrepository/app1",
	}
	status1 := &corev1.ContainerStatus{
		Name:         "app1",
		Ready:        true,
		RestartCount: 0,
	}
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "pod1",
			CreationTimestamp: metav1.Now(),
			Annotations:       map[string]string{"some.test/anno.tation": "123"},
			Labels:            map[string]string{"hello": "world", "tier": "backend"},
		},
		Spec: corev1.PodSpec{
			NodeName:   "worker-2",
			Containers: []corev1.Container{*container1},
		},
		Status: corev1.PodStatus{
			Phase:             "Running",
			ContainerStatuses: []corev1.ContainerStatus{*status1},
		},
	}
	h, _ := setupHandler(t, pod1)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "Invalidnamespace", "pn": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetail(w, r)

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

func TestGetPodDetails_BadRequest_PN(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "pn": "&%$"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetail(w, r)

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
	assert.Contains(t, errorResponse.Message, "invalid pod")
}

func TestGetPodDetails_DNE(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "pn": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetail(w, r)

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
	assert.Equal(t, "pod not found", errorResponse.Message)
}

func TestGetPodDetails_Error(t *testing.T) {
	// arrange
	h, fakeCS := setupHandler(t)
	fakeCS.PrependReactor("get", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "pn": "app1-87745fd6c1-9gd3s"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetail(w, r)

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
	assert.Equal(t, "failed to retrieve pod details", errorResponse.Message)
}

func TestGetEvents_HappyPath(t *testing.T) {
	now := time.Now()
	hourAgo := now.Add(-10 * time.Minute)
	event1 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event1",
			Namespace: "default",
		},
		Type:           "Normal",
		Reason:         "Scheduled",
		Message:        "Successfully assigned default/pod1 to worker-1",
		Count:          1,
		FirstTimestamp: metav1.Time{Time: hourAgo},
		LastTimestamp:  metav1.Time{Time: now},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod1",
			Namespace: "default",
		},
	}
	event2 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event2",
			Namespace: "default",
		},
		Type:           "Normal",
		Reason:         "SuccessfulCreate",
		Message:        "Created pod: pod1",
		Count:          1,
		FirstTimestamp: metav1.Time{Time: hourAgo},
		LastTimestamp:  metav1.Time{Time: now},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod1",
			Namespace: "default",
		},
	}
	event3 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event3",
			Namespace: "othernamespace",
		},
		Type:           "Normal",
		Reason:         "Scheduled",
		Message:        "Successfully assigned othernamespace/pod2 to worker-1",
		Count:          1,
		FirstTimestamp: metav1.Time{Time: hourAgo},
		LastTimestamp:  metav1.Time{Time: now},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod2",
			Namespace: "othernamespace",
		},
	}
	event4 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event4",
			Namespace: "othernamespace",
		},
		Type:           "Normal",
		Reason:         "SuccessfulCreate",
		Message:        "Created pod: pod2",
		Count:          1,
		FirstTimestamp: metav1.Time{Time: hourAgo},
		LastTimestamp:  metav1.Time{Time: now},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod2",
			Namespace: "othernamespace",
		},
	}
	h, _ := setupHandler(t, event1, event2, event3, event4)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListEvents(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var events []types.Event
	err = json.Unmarshal(body, &events)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(events))
	assert.Equal(t, "Normal", events[1].Type)
	assert.Equal(t, "Scheduled", events[0].Reason)
	assert.Equal(t, "Successfully assigned default/pod1 to worker-1", events[0].Message)
	assert.Equal(t, int32(1), events[0].Count)
	assert.True(t, hourAgo.Equal(events[0].FirstTime))
	assert.True(t, now.Equal(events[0].LastTime))
	assert.Equal(t, "Normal", events[0].Type)
	assert.Equal(t, "SuccessfulCreate", events[1].Reason)
	assert.Equal(t, "Created pod: pod1", events[1].Message)
	assert.Equal(t, int32(1), events[1].Count)
	assert.True(t, hourAgo.Equal(events[1].FirstTime))
	assert.True(t, now.Equal(events[1].LastTime))
}

func TestGetEvents_WithFilter(t *testing.T) {
	h, fakeCS := setupHandler(t)
	var capturedFieldSelector string
	fakeCS.PrependReactor("list", "events", func(action k8stesting.Action) (bool, runtime.Object, error) {
		listAction := action.(k8stesting.ListAction)
		capturedFieldSelector = listAction.GetListRestrictions().Fields.String()
		return false, nil, nil
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request?involvedObjectName=pod2", map[string]string{"ns": "default"})
	r.URL.Query()
	w := httptest.NewRecorder()

	// act
	h.ListEvents(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var events []types.Event
	err = json.Unmarshal(body, &events)
	assert.NoError(t, err)
	assert.Equal(t, `involvedObject.name=pod2`, capturedFieldSelector)
}

func TestGetEvents_None(t *testing.T) {
	now := time.Now()
	hourAgo := now.Add(-10 * time.Minute)
	event1 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event1",
			Namespace: "othernamespace",
		},
		Type:           "Normal",
		Reason:         "Scheduled",
		Message:        "Successfully assigned default/pod1 to worker-1",
		Count:          1,
		FirstTimestamp: metav1.Time{Time: hourAgo},
		LastTimestamp:  metav1.Time{Time: now},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod1",
			Namespace: "othernamespace",
		},
	}
	event2 := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "event2",
			Namespace: "othernamespace",
		},
		Type:           "Normal",
		Reason:         "SuccessfulCreate",
		Message:        "Created pod: pod1",
		Count:          1,
		FirstTimestamp: metav1.Time{Time: hourAgo},
		LastTimestamp:  metav1.Time{Time: now},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Name:      "pod1",
			Namespace: "othernamespace",
		},
	}
	h, _ := setupHandler(t, event1, event2)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListEvents(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotNil(t, result.Body)
	body, err := io.ReadAll(result.Body)
	assert.NoError(t, err)
	var events []types.Event
	err = json.Unmarshal(body, &events)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(events))
}

func TestGetEvents_BadRequest_NS(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "Invalidnamepace"})
	w := httptest.NewRecorder()

	// act
	h.ListEvents(w, r)

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

func TestGetEvents_BadRequest_Filter(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request?involvedObjectName=Invalidpodname", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListEvents(w, r)

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
	assert.Contains(t, errorResponse.Message, "invalid object filter")
}

func TestGetEvents_Error(t *testing.T) {
	// arrange
	h, fakeCS := setupHandler(t)
	fakeCS.PrependReactor("list", "events", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("kube api unavailable")
	})
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default"})
	w := httptest.NewRecorder()

	// act
	h.ListEvents(w, r)

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
	assert.Equal(t, "failed to fetch events", errorResponse.Message)
}

func TestValidateNamespaceName(t *testing.T) {
	tests := []struct {
		caseName  string
		name      string
		expectErr bool
	}{
		{caseName: "only alphanumeric", name: "mynamespace"},
		{caseName: "with numbers", name: "mynam35pace"},
		{caseName: "with hypens", name: "my-namespace"},
		{caseName: "with numbers and hyphens", name: "my-nam35pace"},
		{caseName: "only numbers", name: "123"},
		{caseName: "start with number", name: "123mynamespace"},
		{caseName: "end with number", name: "mynamepace123"},
		{caseName: "one", name: "a"},
		{caseName: "two", name: "ab"},
		{caseName: "exactly at cap", name: strings.Repeat("a", 63)},
		{caseName: "over cap", name: strings.Repeat("a", 64), expectErr: true},
		{caseName: "uppercase", name: "myNamespace", expectErr: true},
		{caseName: "whitespace", name: "my namespace", expectErr: true},
		{caseName: "empty string", name: "", expectErr: true},
		{caseName: "start with hyphen", name: "-mynamespace", expectErr: true},
		{caseName: "end with hyphen", name: "mynamespace-", expectErr: true},
		{caseName: "underscore", name: "my_namespace", expectErr: true},
		{caseName: "dot", name: "my.namespace", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.caseName, func(t *testing.T) {
			err := validateNamespaceName(tc.name)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		caseName  string
		name      string
		expectErr bool
	}{
		{caseName: "only alphanumeric", name: "myresource"},
		{caseName: "with numbers", name: "myr35ource"},
		{caseName: "with hypens", name: "my-resource"},
		{caseName: "with numbers and hyphens", name: "my-r35ource"},
		{caseName: "only numbers", name: "123"},
		{caseName: "start with number", name: "123myresource"},
		{caseName: "end with number", name: "myresource123"},
		{caseName: "one", name: "a"},
		{caseName: "two", name: "ab"},
		{caseName: "exactly at cap", name: strings.Repeat("a", 253)},
		{caseName: "over cap", name: strings.Repeat("a", 254), expectErr: true},
		{caseName: "uppercase", name: "myResource", expectErr: true},
		{caseName: "whitespace", name: "my resource", expectErr: true},
		{caseName: "empty string", name: "", expectErr: true},
		{caseName: "start with hyphen", name: "-myresource", expectErr: true},
		{caseName: "end with hyphen", name: "myresource-", expectErr: true},
		{caseName: "underscore", name: "my_resource", expectErr: true},
		{caseName: "dot", name: "my.resource"},
	}

	for _, tc := range tests {
		t.Run(tc.caseName, func(t *testing.T) {
			err := validateResourceName(tc.name)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestParseTailLines(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		expected  int64
		expectErr bool
	}{
		{name: "empty defaults to 100", raw: "", expected: 100},
		{name: "valid value passes through", raw: "250", expected: 250},
		{name: "over cap clamps to 1000", raw: "5000", expected: 1000},
		{name: "exactly at cap", raw: "1000", expected: 1000},
		{name: "invalid string", raw: "notanumber", expectErr: true},
		{name: "zero", raw: "0", expectErr: true},
		{name: "negative", raw: "-5", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseTailLines(tc.raw)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
