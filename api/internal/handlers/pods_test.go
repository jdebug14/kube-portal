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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

func TestListPods_HappyPath(t *testing.T) {
	// arrange
	now := metav1.Now()
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
	otherPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "kube-system",
			Name:              "coredns",
			CreationTimestamp: now,
		},
	}
	h, _ := setupHandler(t, pod1, otherPod)
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
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetails(w, r)

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
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetails(w, r)

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

func TestGetPodDetails_BadRequest_namespace(t *testing.T) {
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
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "Invalidnamespace", "name": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetails(w, r)

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

func TestGetPodDetails_BadRequest_name(t *testing.T) {
	// arrange
	h, _ := setupHandler(t)
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "&%$"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetails(w, r)

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
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "pod1"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetails(w, r)

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
	r := newRequestWithParams(t, http.MethodGet, "/some/test/request", map[string]string{"ns": "default", "name": "app1-87745fd6c1-9gd3s"})
	w := httptest.NewRecorder()

	// act
	h.GetPodDetails(w, r)

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
	assert.Equal(t, "failed to get pod details", errorResponse.Message)
}
