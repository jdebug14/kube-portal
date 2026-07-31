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
