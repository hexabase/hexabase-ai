//go:build unit

package tekton_test

import (
	"context"
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	"github.com/stretchr/testify/assert"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"

	"github.com/hexabase/hexabase-ai/api/internal/cicd/tekton"
)

func TestTektonProvider_RunPipeline(t *testing.T) {
	t.Run("should create a Tekton PipelineRun successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		tektonClient := fake.NewSimpleClientset()
		
		provider := tekton.NewTektonProvider(tektonClient)

		projectID := "my-awesome-project"
		namespace := "workspace-ns-123" // This would be derived from the projectID in a real scenario
		runOptions := domain.RunOptions{
			RepositoryURL: "https://github.com/hexabase/hexabase-app.git",
			Branch:        "feature/new-ui",
			AppName:       "my-app",
			ImageName:     "hexabase/my-app:feature-new-ui",
		}

		// Act
		result, err := provider.RunPipeline(ctx, projectID, namespace, runOptions)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.RunID)
		assert.Equal(t, "Pending", result.Status)
		assert.Contains(t, result.Message, "Tekton PipelineRun created")

		// Verify that the PipelineRun was created in the fake clientset
		pr, err := tektonClient.TektonV1().PipelineRuns(namespace).Get(ctx, result.RunID, metav1.GetOptions{})
		assert.NoError(t, err)
		assert.NotNil(t, pr)

		// Check some key values of the created PipelineRun
		assert.Equal(t, "build-and-push-pipeline", pr.Spec.PipelineRef.Name)
		assert.Contains(t, pr.Name, "my-app-run-")
		
		expectedParams := map[string]string{
			"repo-url":    runOptions.RepositoryURL,
			"revision":    runOptions.Branch,
			"image-ref":   runOptions.ImageName,
			"app-name":    runOptions.AppName,
		}

		for _, p := range pr.Spec.Params {
			val, ok := expectedParams[p.Name]
			assert.True(t, ok, "Unexpected parameter found: %s", p.Name)
			assert.Equal(t, val, p.Value.StringVal)
		}
	})
}

func TestTektonProvider_GetRunStatus(t *testing.T) {
	t.Run("should get the status of a PipelineRun", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		namespace := "workspace-ns-123"
		runID := "my-app-run-succeeded"

		// Create a fake PipelineRun object with a "Succeeded" status condition
		fakePipelineRun := &v1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runID,
				Namespace: namespace,
			},
			Status: v1.PipelineRunStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   apis.ConditionSucceeded,
							Status: "True",
							Reason: "Completed",
						},
					},
				},
			},
		}
		
		tektonClient := fake.NewSimpleClientset(fakePipelineRun)
		provider := tekton.NewTektonProvider(tektonClient)

		// Act
		result, err := provider.GetRunStatus(ctx, "my-project", namespace, runID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, runID, result.RunID)
		assert.Equal(t, "Succeeded", result.Status)
	})

	t.Run("should return 'Failed' for a failed PipelineRun", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		namespace := "workspace-ns-123"
		runID := "my-app-run-failed"

		fakePipelineRun := &v1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{Name: runID, Namespace: namespace},
			Status: v1.PipelineRunStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   apis.ConditionSucceeded,
							Status: "False",
							Reason: "Error",
						},
					},
				},
			},
		}

		tektonClient := fake.NewSimpleClientset(fakePipelineRun)
		provider := tekton.NewTektonProvider(tektonClient)

		// Act
		result, err := provider.GetRunStatus(ctx, "my-project", namespace, runID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Failed", result.Status)
	})

	t.Run("should return 'Running' for a running PipelineRun", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		namespace := "workspace-ns-123"
		runID := "my-app-run-running"

		fakePipelineRun := &v1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{Name: runID, Namespace: namespace},
			Status: v1.PipelineRunStatus{
				Status: duckv1.Status{
					Conditions: duckv1.Conditions{
						{
							Type:   apis.ConditionSucceeded,
							Status: "Unknown",
							Reason: "Running",
						},
					},
				},
			},
		}

		tektonClient := fake.NewSimpleClientset(fakePipelineRun)
		provider := tekton.NewTektonProvider(tektonClient)

		// Act
		result, err := provider.GetRunStatus(ctx, "my-project", namespace, runID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Running", result.Status)
	})
} 