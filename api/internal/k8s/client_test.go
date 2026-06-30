package k8s

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMapContainers_HappyPath(t *testing.T) {
	// arrange
	name1 := "app1"
	now := metav1.Now().Time
	app1 := v1.Container{
		Name:  name1,
		Image: "myrepository/" + name1,
	}
	containerStatus1 := v1.ContainerStatus{
		Name:         name1,
		Ready:        false,
		RestartCount: 17,
		LastTerminationState: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 1,
				Reason:   "Error",
				FinishedAt: metav1.Time{
					Time: now,
				},
			},
		},
	}

	// act
	results := mapContainers([]v1.Container{app1}, []v1.ContainerStatus{containerStatus1})

	// assert
	if assert.NotNil(t, results, "results") {
		assert.Equal(t, 1, len(results), "results size")
		r1 := results[0]
		assert.Equal(t, name1, r1.Name, "container name")
		assert.Equal(t, "myrepository/"+name1, r1.Image, "container image")
		assert.False(t, r1.Ready, "container ready")
		assert.Equal(t, int32(17), r1.Restarts, "container restarts")
		assert.Equal(t, "Error", *r1.LastExitReason, "container last exit reason")
		assert.True(t, r1.LastExitTime.Equal(now), "container last exit time")
	}
}

func TestMapContainers_NeverTerminated(t *testing.T) {
	// arrange
	name1 := "app1"
	app1 := v1.Container{
		Name:  name1,
		Image: "myrepository/" + name1,
	}
	containerStatus1 := v1.ContainerStatus{
		Name:         name1,
		Ready:        true,
		RestartCount: 0,
	}

	// act
	results := mapContainers([]v1.Container{app1}, []v1.ContainerStatus{containerStatus1})

	// assert
	if assert.NotNil(t, results, "results") {
		assert.Equal(t, 1, len(results), "results size")
		r1 := results[0]
		assert.Equal(t, name1, r1.Name, "container name")
		assert.Equal(t, "myrepository/"+name1, r1.Image, "container image")
		assert.True(t, r1.Ready, "container ready")
		assert.Equal(t, int32(0), r1.Restarts, "container restarts")
		assert.Nil(t, r1.LastExitReason, "container last exit reason")
		assert.Nil(t, r1.LastExitTime, "container last exit time")
	}
}

func TestMapContainers_MultipleContainers(t *testing.T) {
	// arrange
	now := metav1.Now().Time
	name1 := "app1"
	app1 := v1.Container{
		Name:  name1,
		Image: "myrepository/" + name1,
	}
	containerStatus1 := v1.ContainerStatus{
		Name:         name1,
		Ready:        true,
		RestartCount: 0,
		LastTerminationState: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 0,
				Reason:   "Completed",
				FinishedAt: metav1.Time{
					Time: now,
				},
			},
		},
	}
	name2 := "app2"
	app2 := v1.Container{
		Name:  name2,
		Image: "myrepository/" + name2,
	}
	containerStatus2 := v1.ContainerStatus{
		Name:         name2,
		Ready:        false,
		RestartCount: 17,
		LastTerminationState: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 1,
				Reason:   "Error",
				FinishedAt: metav1.Time{
					Time: now,
				},
			},
		},
	}

	// act
	results := mapContainers([]v1.Container{app1, app2}, []v1.ContainerStatus{containerStatus1, containerStatus2})

	// assert
	if assert.NotNil(t, results) {
		assert.Equal(t, 2, len(results), "results size")
		r1 := results[0]
		assert.Equal(t, name1, r1.Name, "app1 name")
		assert.Equal(t, "myrepository/"+name1, r1.Image, "app1 image")
		assert.True(t, r1.Ready, "app1 ready")
		assert.Equal(t, int32(0), r1.Restarts, "app1 restarts")
		assert.Equal(t, "Completed", *r1.LastExitReason, "app1 last exit reason")
		assert.True(t, r1.LastExitTime.Equal(now), "app1 last exit time")
		r2 := results[1]
		assert.Equal(t, name2, r2.Name, "app2 name")
		assert.Equal(t, "myrepository/"+name2, r2.Image, "app2 image")
		assert.False(t, r2.Ready, "app2 ready")
		assert.Equal(t, int32(17), r2.Restarts, "app2 restarts")
		assert.Equal(t, "Error", *r2.LastExitReason, "app2 last exit reason")
		assert.True(t, r2.LastExitTime.Equal(now), "app2 last exit time")
	}
}

func TestMapContainers_Empty(t *testing.T) {
	// arrange

	// act
	results := mapContainers([]v1.Container{}, []v1.ContainerStatus{})

	// assert
	if assert.NotNil(t, results, "results") {
		assert.Equal(t, 0, len(results), "results size")
	}
}

func TestMapContainers_EmptyContainers(t *testing.T) {
	// arrange
	containerStatus1 := v1.ContainerStatus{
		Name:         "app1",
		Ready:        true,
		RestartCount: 0,
		LastTerminationState: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 0,
				Reason:   "Completed",
				FinishedAt: metav1.Time{
					Time: time.Now(),
				},
			},
		},
	}

	// act
	results := mapContainers([]v1.Container{}, []v1.ContainerStatus{containerStatus1})

	// assert
	if assert.NotNil(t, results, "results") {
		assert.Equal(t, 0, len(results), "results size")
	}
}

func TestMapContainers_EmptyContainerStatuses(t *testing.T) {
	// arrange
	name1 := "app1"
	app1 := v1.Container{
		Name:  name1,
		Image: "myrepository/" + name1,
	}

	// act
	results := mapContainers([]v1.Container{app1}, []v1.ContainerStatus{})

	// assert
	if assert.NotNil(t, results, "results") {
		assert.Equal(t, 1, len(results), "results size")
		r1 := results[0]
		assert.Equal(t, name1, r1.Name, "container name")
		assert.Equal(t, "myrepository/"+name1, r1.Image, "container image")
		assert.False(t, r1.Ready, "container ready")
		assert.Equal(t, int32(0), r1.Restarts, "container restarts")
		assert.Nil(t, r1.LastExitReason, "last exit reason")
		assert.Nil(t, r1.LastExitTime, "last exit time")
	}
}

func TestMapContainers_NonMatchingContainerStatus(t *testing.T) {
	// arrange
	now := time.Now()
	name1 := "app1"
	name2 := "app2"
	app1 := v1.Container{
		Name:  name1,
		Image: "myrepository/" + name1,
	}
	containerStatus1 := v1.ContainerStatus{
		Name:         name2,
		Ready:        true,
		RestartCount: 0,
		LastTerminationState: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 0,
				Reason:   "Completed",
				FinishedAt: metav1.Time{
					Time: now,
				},
			},
		},
	}

	// act
	results := mapContainers([]v1.Container{app1}, []v1.ContainerStatus{containerStatus1})

	// assert
	if assert.NotNil(t, results, "results") {
		assert.Equal(t, 1, len(results), "results size")
		r1 := results[0]
		assert.Equal(t, name1, r1.Name, "container name")
		assert.Equal(t, "myrepository/"+name1, r1.Image, "container image")
		assert.False(t, r1.Ready, "container ready")
		assert.Equal(t, int32(0), r1.Restarts)
		assert.Nil(t, r1.LastExitReason)
		assert.Nil(t, r1.LastExitTime)
	}
}
