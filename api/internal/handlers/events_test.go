package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jdebug14/kube-portal/internal/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

func TestListEvents_HappyPath(t *testing.T) {
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

func TestListEvents_WithFilter(t *testing.T) {
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

func TestListEvents_None(t *testing.T) {
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

func TestListEvents_BadRequest_namespace(t *testing.T) {
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

func TestListEvents_BadRequest_Filter(t *testing.T) {
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

func TestListEvents_Error(t *testing.T) {
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
