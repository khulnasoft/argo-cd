package reporter

import (
	"fmt"
	"testing"
	"time"

	"github.com/argoproj/gitops-engine/pkg/health"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/gitops-engine/pkg/sync/common"
	"github.com/stretchr/testify/assert"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

func TestParseResourceSyncResultErrors(t *testing.T) {
	t.Run("Resource of app contain error that contain comma", func(t *testing.T) {
		errors := parseResourceSyncResultErrors(&v1alpha1.ResourceStatus{
			Group:     "group",
			Kind:      "kind",
			Namespace: "namespace",
			Name:      "name",
		}, &v1alpha1.OperationState{
			SyncResult: &v1alpha1.SyncOperationResult{
				Resources: v1alpha1.ResourceResults{
					{
						Group:     "group",
						Kind:      "kind",
						Namespace: "namespace",
						Name:      "name",
						SyncPhase: common.SyncPhaseSync,
						Message:   "error message, with comma",
						HookPhase: common.OperationFailed,
					},
					{
						Group:     "group",
						Kind:      "kind",
						Namespace: "namespace",
						Name:      "name-2",
						SyncPhase: common.SyncPhaseSync,
					},
				},
				Revision: "123",
				Source:   v1alpha1.ApplicationSource{},
			},
		})

		assert.Len(t, errors, 1)
		assert.Equal(t, "error message, with comma", errors[0].Message)
		assert.Equal(t, "sync", errors[0].Type)
		assert.Equal(t, "error", errors[0].Level)
	})
	t.Run("Resource of app contain error", func(t *testing.T) {
		errors := parseResourceSyncResultErrors(&v1alpha1.ResourceStatus{
			Group:     "group",
			Kind:      "kind",
			Namespace: "namespace",
			Name:      "name",
		}, &v1alpha1.OperationState{
			SyncResult: &v1alpha1.SyncOperationResult{
				Resources: v1alpha1.ResourceResults{
					{
						Group:     "group",
						Kind:      "kind",
						Namespace: "namespace",
						Name:      "name",
						SyncPhase: common.SyncPhaseSync,
						Message:   "error message",
						HookPhase: common.OperationFailed,
					},
					{
						Group:     "group",
						Kind:      "kind",
						Namespace: "namespace",
						Name:      "name-2",
						SyncPhase: common.SyncPhaseSync,
					},
				},
				Revision: "123",
				Source:   v1alpha1.ApplicationSource{},
			},
		})

		assert.Len(t, errors, 1)
		assert.Equal(t, "error message", errors[0].Message)
		assert.Equal(t, "sync", errors[0].Type)
		assert.Equal(t, "error", errors[0].Level)
	})
	t.Run("Resource of app not contain error", func(t *testing.T) {
		errors := parseResourceSyncResultErrors(&v1alpha1.ResourceStatus{
			Group:     "group",
			Kind:      "kind",
			Namespace: "namespace",
			Name:      "name",
		}, &v1alpha1.OperationState{
			SyncResult: &v1alpha1.SyncOperationResult{
				Resources: v1alpha1.ResourceResults{
					{
						Group:     "group",
						Kind:      "kind",
						Namespace: "namespace",
						Name:      "name",
						SyncPhase: common.SyncPhaseSync,
						Message:   "error message",
						HookPhase: common.OperationSucceeded,
					},
					{
						Group:     "group",
						Kind:      "kind",
						Namespace: "namespace",
						Name:      "name-2",
						SyncPhase: common.SyncPhaseSync,
					},
				},
				Revision: "123",
				Source:   v1alpha1.ApplicationSource{},
			},
		})

		assert.Empty(t, errors)
	})
	t.Run("App of app contain error", func(t *testing.T) {
		errors := parseApplicationSyncResultErrors(&v1alpha1.OperationState{
			Phase:   common.OperationError,
			Message: "error message",
			SyncResult: &v1alpha1.SyncOperationResult{
				Revision: "123",
				Source:   v1alpha1.ApplicationSource{},
			},
		})

		assert.Len(t, errors, 1)
		assert.Equal(t, "error message", errors[0].Message)
		assert.Equal(t, "sync", errors[0].Type)
		assert.Equal(t, "error", errors[0].Level)
	})
}

func TestGetConditionLevel(t *testing.T) {
	errorsTypes := []v1alpha1.ApplicationCondition{
		{Type: v1alpha1.ApplicationConditionDeletionError},
		{Type: v1alpha1.ApplicationConditionInvalidSpecError},
		{Type: v1alpha1.ApplicationConditionComparisonError},
		{Type: v1alpha1.ApplicationConditionSyncError},
		{Type: v1alpha1.ApplicationConditionUnknownError},
	}
	for _, errorAppCondition := range errorsTypes {
		t.Run("parses as error "+errorAppCondition.Type, func(t *testing.T) {
			level := getConditionLevel(errorAppCondition)
			assert.Equal(t, "error", level)
		})
	}

	warningTypes := []v1alpha1.ApplicationCondition{
		{Type: v1alpha1.ApplicationConditionSharedResourceWarning},
		{Type: v1alpha1.ApplicationConditionRepeatedResourceWarning},
		{Type: v1alpha1.ApplicationConditionExcludedResourceWarning},
		{Type: v1alpha1.ApplicationConditionOrphanedResourceWarning},
	}
	for _, warningAppCondition := range warningTypes {
		t.Run("parses as warning "+warningAppCondition.Type, func(t *testing.T) {
			level := getConditionLevel(warningAppCondition)
			assert.Equal(t, "warning", level)
		})
	}
}

func TestParseApplicationSyncResultErrorsFromConditions(t *testing.T) {
	t.Run("conditions exists", func(t *testing.T) {
		errors := parseApplicationSyncResultErrorsFromConditions(v1alpha1.ApplicationStatus{
			Conditions: []v1alpha1.ApplicationCondition{
				{
					Type:    v1alpha1.ApplicationConditionSyncError,
					Message: "error message",
				},
			},
		})

		assert.Len(t, errors, 1)
		assert.Equal(t, "error message", errors[0].Message)
		assert.Equal(t, "sync", errors[0].Type)
		assert.Equal(t, "error", errors[0].Level)
	})

	t.Run("warning exists", func(t *testing.T) {
		errors := parseApplicationSyncResultErrorsFromConditions(v1alpha1.ApplicationStatus{
			Conditions: []v1alpha1.ApplicationCondition{
				{
					Type:    v1alpha1.ApplicationConditionOrphanedResourceWarning,
					Message: "Application has 8 orphaned resources",
				},
			},
		})

		assert.Len(t, errors, 1)
		assert.Equal(t, "Application has 8 orphaned resources", errors[0].Message)
		assert.Equal(t, "sync", errors[0].Type)
		assert.Equal(t, "warning", errors[0].Level)
	})

	t.Run("conditions erorr replaced with sync result errors", func(t *testing.T) {
		errors := parseApplicationSyncResultErrorsFromConditions(v1alpha1.ApplicationStatus{
			Conditions: []v1alpha1.ApplicationCondition{
				{
					Type:    v1alpha1.ApplicationConditionSyncError,
					Message: syncTaskUnsuccessfullErrorMessage,
				},
			},
			OperationState: &v1alpha1.OperationState{
				SyncResult: &v1alpha1.SyncOperationResult{
					Resources: v1alpha1.ResourceResults{
						&v1alpha1.ResourceResult{
							Kind:      "Job",
							Name:      "some-job",
							Message:   "job failed",
							HookPhase: common.OperationFailed,
						},
						&v1alpha1.ResourceResult{
							Kind:    "Pod",
							Name:    "some-pod",
							Message: "pod failed",
							Status:  common.ResultCodeSyncFailed,
						},
						&v1alpha1.ResourceResult{
							Kind:      "Job",
							Name:      "some-succeeded-hook",
							Message:   "job succeeded",
							HookPhase: common.OperationSucceeded,
						},
						&v1alpha1.ResourceResult{
							Kind:    "Pod",
							Name:    "synced-pod",
							Message: "pod synced",
							Status:  common.ResultCodeSynced,
						},
					},
				},
			},
		})

		assert.Len(t, errors, 2)
		assert.Equal(t, fmt.Sprintf("Resource %s(%s): \n %s", "Job", "some-job", "job failed"), errors[0].Message)
		assert.Equal(t, "sync", errors[0].Type)
		assert.Equal(t, "error", errors[0].Level)
		assert.Equal(t, fmt.Sprintf("Resource %s(%s): \n %s", "Pod", "some-pod", "pod failed"), errors[1].Message)
		assert.Equal(t, "sync", errors[1].Type)
		assert.Equal(t, "error", errors[1].Level)
	})
}

func TestParseAggregativeHealthErrors(t *testing.T) {
	t.Run("application tree is nil", func(t *testing.T) {
		errs := parseAggregativeHealthErrors(&v1alpha1.ResourceStatus{
			Group:     "group",
			Kind:      "application",
			Namespace: "namespace",
			Name:      "name",
		}, nil, false)
		assert.Empty(t, errs)
	})

	t.Run("should set sourceReference", func(t *testing.T) {
		rsName := "test-deployment"
		ns := "test"
		errMessage := "backoff pulling image test/test:0.1"
		rsRef := v1alpha1.ResourceRef{
			Group:     "g",
			Version:   "v",
			Kind:      "ReplicaSet",
			Name:      rsName + "1",
			Namespace: ns,
		}

		deployRef := v1alpha1.ResourceRef{
			Group:     "g",
			Version:   "v",
			Kind:      "Deployment",
			Name:      rsName,
			Namespace: ns,
		}

		appTree := v1alpha1.ApplicationTree{
			Nodes: []v1alpha1.ResourceNode{
				{ // Pod
					Health: &v1alpha1.HealthStatus{
						Status:  health.HealthStatusDegraded,
						Message: errMessage,
					},
					ResourceRef: v1alpha1.ResourceRef{
						Group:     "g",
						Version:   "v",
						Kind:      "Pod",
						Name:      rsName + "1-3n235j5",
						Namespace: ns,
					},
					ParentRefs: []v1alpha1.ResourceRef{rsRef},
					CreatedAt: &metav1.Time{
						Time: time.Now(),
					},
				},
				{ // ReplicaSet
					Health: &v1alpha1.HealthStatus{
						Status:  health.HealthStatusProgressing,
						Message: "",
					},
					ResourceRef: rsRef,
					ParentRefs:  []v1alpha1.ResourceRef{deployRef},
					CreatedAt: &metav1.Time{
						Time: time.Now(),
					},
				},
				{ // Deployment
					Health: &v1alpha1.HealthStatus{
						Status:  health.HealthStatusDegraded,
						Message: "",
					},
					ResourceRef: deployRef,
					ParentRefs:  []v1alpha1.ResourceRef{},
					CreatedAt: &metav1.Time{
						Time: time.Now(),
					},
				},
			},
		}

		errs := parseAggregativeHealthErrors(&v1alpha1.ResourceStatus{
			Group:     deployRef.Group,
			Version:   deployRef.Version,
			Kind:      deployRef.Kind,
			Name:      deployRef.Name,
			Namespace: deployRef.Namespace,
		}, &appTree, true)
		assert.Len(t, errs, 1)
		assert.Equal(t, errMessage, errs[0].Message)
		assert.NotNil(t, errs[0].SourceReference)
		assert.Equal(t, deployRef.Name, errs[0].SourceReference.Name)
	})

	t.Run("should build error from root node if it's degraded while no one of child in degraded health state", func(t *testing.T) {
		rsName := "test-deployment"
		ns := "test"
		errMessage := "backoff pulling image test/test:0.1"
		replicaSetRef := v1alpha1.ResourceRef{
			Group:     "g",
			Version:   "v",
			Kind:      "ReplicaSet",
			Name:      rsName + "1",
			Namespace: ns,
		}

		deploymentRef := v1alpha1.ResourceRef{
			Group:     "g",
			Version:   "v",
			Kind:      "Deployment",
			Name:      rsName,
			Namespace: ns,
		}

		appTree := v1alpha1.ApplicationTree{
			Nodes: []v1alpha1.ResourceNode{
				{ // Pod
					ResourceRef: v1alpha1.ResourceRef{
						Group:     "g",
						Version:   "v",
						Kind:      "Pod",
						Name:      rsName + "1-3n235j5",
						Namespace: ns,
					},
					Health: &v1alpha1.HealthStatus{
						Status:  health.HealthStatusProgressing,
						Message: "some error of pod",
					},
					ParentRefs: []v1alpha1.ResourceRef{replicaSetRef},
					CreatedAt: &metav1.Time{
						Time: time.Now(),
					},
				},
				{ // ReplicaSet
					ResourceRef: replicaSetRef,
					Health: &v1alpha1.HealthStatus{
						Status:  health.HealthStatusProgressing,
						Message: "",
					},
					ParentRefs: []v1alpha1.ResourceRef{deploymentRef},
					CreatedAt: &metav1.Time{
						Time: time.Now(),
					},
				},
				{ // Deployment
					ResourceRef: deploymentRef,
					Health: &v1alpha1.HealthStatus{
						Status:  health.HealthStatusDegraded,
						Message: errMessage,
					},
					ParentRefs: []v1alpha1.ResourceRef{},
					CreatedAt: &metav1.Time{
						Time: time.Now(),
					},
				},
			},
		}

		errs := parseAggregativeHealthErrors(&v1alpha1.ResourceStatus{
			Group:     deploymentRef.Group,
			Version:   deploymentRef.Version,
			Kind:      deploymentRef.Kind,
			Name:      deploymentRef.Name,
			Namespace: deploymentRef.Namespace,
		}, &appTree, true)
		assert.Len(t, errs, 1)
		assert.Equal(t, errMessage, errs[0].Message)
		assert.NotNil(t, errs[0].SourceReference)
		assert.Equal(t, deploymentRef.Name, errs[0].SourceReference.Name)
	})
}
