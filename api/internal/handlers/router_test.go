package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jdebug14/kube-portal/internal/k8s"
	"github.com/jdebug14/kube-portal/internal/types"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func setupRouter(t *testing.T, objects ...runtime.Object) *chi.Mux {
	t.Helper()
	fakeCS := fake.NewSimpleClientset(objects...)
	k8sClient := k8s.NewClientFromInterface(fakeCS, slog.Default())
	h := NewHandler(k8sClient, slog.Default())
	return NewRouter(h)
}

func TestRouter_HealthCheck(t *testing.T) {
	// arrange
	router := setupRouter(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	// act
	router.ServeHTTP(w, r)

	//assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		if assert.NoError(t, err) {
			var health map[string]string
			err = json.Unmarshal(body, &health)
			if assert.NoError(t, err) {
				assert.Equal(t, 1, len(health))
				assert.Equal(t, "ok", health["status"])
			}
		}
	}
}

func TestRouter_ListNamespaces(t *testing.T) {
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
	router := setupRouter(t, defaultNamespace)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		if assert.NoError(t, err) {
			var namespaces []types.Namespace
			err = json.Unmarshal(body, &namespaces)
			if assert.NoError(t, err) {
				assert.Equal(t, 1, len(namespaces))
				assert.Equal(t, "default", namespaces[0].Name)
			}
		}
	}
}

func TestRouter_ListDeployments(t *testing.T) {
	// arrange
	now := metav1.Now()
	replicas := int32(3)
	deployment1 := &appsv1.Deployment{
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
	router := setupRouter(t, deployment1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/deployments", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	body, err := io.ReadAll(result.Body)
	if assert.NoError(t, err) {
		var deployments []types.Deployment
		err = json.Unmarshal(body, &deployments)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, len(deployments))
			assert.Equal(t, "app1", deployments[0].Name)
		}
	}
}

func TestRouter_GetDeploymentDetails(t *testing.T) {
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

	router := setupRouter(t, deployment1, replicaset1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/deployments/deployment1", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		if assert.NoError(t, err) {
			var deploymentDetails types.DeploymentDetails
			err = json.Unmarshal(body, &deploymentDetails)
			if assert.NoError(t, err) {
				assert.Equal(t, "deployment1", deploymentDetails.Name)
			}
		}
	}
}

func TestRouter_ListPods(t *testing.T) {
	// arrange
	now := metav1.Now()
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "pod1",
			CreationTimestamp: now,
		},
		Spec: corev1.PodSpec{
			NodeName: "worker-1",
		},
		Status: corev1.PodStatus{
			Phase: "Running",
		},
	}
	router := setupRouter(t, pod1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		if assert.NoError(t, err) {
			var pods []types.Pod
			err = json.Unmarshal(body, &pods)
			if assert.NoError(t, err) {
				assert.Equal(t, 1, len(pods))
				assert.Equal(t, "pod1", pods[0].Name)
			}
		}
	}
}

func TestRouter_GetPodDetails(t *testing.T) {
	// arrange
	now := metav1.Now()
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "pod1",
			CreationTimestamp: now,
		},
		Spec: corev1.PodSpec{
			NodeName: "worker-1",
		},
		Status: corev1.PodStatus{
			Phase: "Running",
		},
	}
	router := setupRouter(t, pod1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/pod1", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		if assert.NoError(t, err) {
			var pod types.Pod
			err = json.Unmarshal(body, &pod)
			if assert.NoError(t, err) {
				assert.Equal(t, "pod1", pod.Name)
			}
		}
	}
}

func TestRouter_ListEvents(t *testing.T) {
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
	router := setupRouter(t, event1, event2)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/events", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	result := w.Result()
	assert.Equal(t, http.StatusOK, result.StatusCode)
	if assert.NotNil(t, result.Body) {
		body, err := io.ReadAll(result.Body)
		if assert.NoError(t, err) {
			var events []types.Event
			err = json.Unmarshal(body, &events)
			if assert.NoError(t, err) {
				assert.Equal(t, 2, len(events))
			}
		}
	}
}

func TestRouter_ListEvents_WithFilter(t *testing.T) {
	// arrange
	fakeCS := fake.NewSimpleClientset()
	var capturedFieldSelector string
	fakeCS.PrependReactor("list", "events", func(action k8stesting.Action) (bool, runtime.Object, error) {
		listAction := action.(k8stesting.ListAction)
		capturedFieldSelector = listAction.GetListRestrictions().Fields.String()
		return false, nil, nil
	})
	k8sClient := k8s.NewClientFromInterface(fakeCS, slog.Default())
	router := NewRouter(NewHandler(k8sClient, slog.Default()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/events?involvedObjectName=pod1", nil)
	w := httptest.NewRecorder()

	// act
	router.ServeHTTP(w, req)

	// assert
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "involvedObject.name=pod1,involvedObject.namespace=default", capturedFieldSelector)
}

func TestRouter_GetLogs(t *testing.T) {
	// arrange
	fakeCS := fake.NewSimpleClientset()
	var capturedOpts *v1.PodLogOptions
	fakeCS.PrependReactor("get", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil // not a logs request, let other handling proceed
		}
		genericAction, ok := action.(k8stesting.GenericAction)
		if ok {
			if opts, ok := genericAction.GetValue().(*corev1.PodLogOptions); ok {
				capturedOpts = opts
			}
		}
		return true, &runtime.Unknown{Raw: []byte("my logs")}, nil
	})
	k8sClient := k8s.NewClientFromInterface(fakeCS, slog.Default())
	router := NewRouter(NewHandler(k8sClient, slog.Default()))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/pod1/logs", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	if assert.NotNil(t, w.Result().Body) {
		body, err := io.ReadAll(w.Result().Body)
		if assert.NoError(t, err) {
			logs := string(body)
			assert.Equal(t, "my logs", logs)
			assert.Equal(t, "", capturedOpts.Container)
			assert.Equal(t, int64(100), *capturedOpts.TailLines)
		}
	}
}

func TestRouter_GetLogs_withContainer(t *testing.T) {
	// arrange
	containerName := "container1"
	fakeCS := fake.NewSimpleClientset()
	var capturedOpts *v1.PodLogOptions
	fakeCS.PrependReactor("get", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil // not a logs request, let other handling proceed
		}
		genericAction, ok := action.(k8stesting.GenericAction)
		if ok {
			if opts, ok := genericAction.GetValue().(*corev1.PodLogOptions); ok {
				capturedOpts = opts
				if opts.Container == containerName {
					return true, &runtime.Unknown{Raw: []byte("my logs")}, nil
				}

			}
		}
		return false, nil, nil
	})
	k8sClient := k8s.NewClientFromInterface(fakeCS, slog.Default())
	router := NewRouter(NewHandler(k8sClient, slog.Default()))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/pod1/logs?container="+containerName, nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	if assert.NotNil(t, w.Result().Body) {
		body, err := io.ReadAll(w.Result().Body)
		if assert.NoError(t, err) {
			logs := string(body)
			assert.Equal(t, "my logs", logs)
			assert.Equal(t, containerName, capturedOpts.Container)
			assert.Equal(t, int64(100), *capturedOpts.TailLines)
		}
	}
}

func TestRouter_GetLogs_withTailLines(t *testing.T) {
	// arrange
	fakeCS := fake.NewSimpleClientset()
	var capturedOpts *v1.PodLogOptions
	fakeCS.PrependReactor("get", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil // not a logs request, let other handling proceed
		}
		genericAction, ok := action.(k8stesting.GenericAction)
		if ok {
			if opts, ok := genericAction.GetValue().(*corev1.PodLogOptions); ok {
				capturedOpts = opts
			}
		}
		return true, &runtime.Unknown{Raw: []byte("my logs")}, nil
	})
	k8sClient := k8s.NewClientFromInterface(fakeCS, slog.Default())
	router := NewRouter(NewHandler(k8sClient, slog.Default()))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/default/pods/pod1/logs?tailLines=10", nil)

	// act
	router.ServeHTTP(w, r)

	// assert
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	if assert.NotNil(t, w.Result().Body) {
		body, err := io.ReadAll(w.Result().Body)
		if assert.NoError(t, err) {
			logs := string(body)
			assert.Equal(t, "my logs", logs)
			assert.Equal(t, "", capturedOpts.Container)
			assert.Equal(t, int64(10), *capturedOpts.TailLines)
		}
	}
}
