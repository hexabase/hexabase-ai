package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	"github.com/stretchr/testify/assert"
)

func TestLangfuseMonitor_CreateTrace(t *testing.T) {
	t.Run("successful trace creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/public/traces", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			// Verify auth headers
			assert.NotEmpty(t, r.Header.Get("X-Langfuse-Public-Key"))
			assert.NotEmpty(t, r.Header.Get("X-Langfuse-Secret-Key"))
			
			// Parse request body
			var reqBody map[string]any
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, "test-trace", reqBody["name"])
			assert.Equal(t, "user-123", reqBody["userId"])
			
			// Return success response
			response := map[string]any{
				"id": reqBody["id"],
				"success": true,
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		trace := &domain.Trace{
			ID:        uuid.New().String(),
			Name:      "test-trace",
			UserID:    "user-123",
			SessionID: "session-456",
			Timestamp: time.Now(),
			Tags:      []string{"test", "unit-test"},
			Metadata: map[string]any{
				"environment": "testing",
			},
		}
		
		// Act
		err := monitor.CreateTrace(ctx, trace)
		
		// Assert
		assert.NoError(t, err)
	})
	
	t.Run("server error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "internal server error",
			})
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		trace := &domain.Trace{
			ID:    uuid.New().String(),
			Name:  "test-trace",
		}
		
		// Act
		err := monitor.CreateTrace(ctx, trace)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
}

func TestLangfuseMonitor_CreateGeneration(t *testing.T) {
	t.Run("successful generation creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/public/generations", r.URL.Path)
			
			var reqBody map[string]any
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, "llama2:7b", reqBody["model"])
			assert.Equal(t, "test-generation", reqBody["name"])
			assert.NotNil(t, reqBody["input"])
			assert.NotNil(t, reqBody["modelParameters"])
			
			response := map[string]any{
				"id": reqBody["id"],
				"success": true,
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		generation := &domain.Generation{
			ID:      uuid.New().String(),
			TraceID: "trace-123",
			Name:    "test-generation",
			Model:   "llama2:7b",
			ModelParameters: map[string]any{
				"temperature": 0.7,
				"max_tokens": 100,
			},
			Input: map[string]string{
				"prompt": "Hello, world!",
			},
			StartTime: time.Now(),
		}
		
		// Act
		err := monitor.CreateGeneration(ctx, generation)
		
		// Assert
		assert.NoError(t, err)
	})
}

func TestLangfuseMonitor_UpdateGeneration(t *testing.T) {
	t.Run("successful generation update", func(t *testing.T) {
		generationID := uuid.New().String()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PATCH", r.Method)
			assert.Equal(t, "/api/public/generations/"+generationID, r.URL.Path)
			
			var reqBody map[string]any
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.NotNil(t, reqBody["output"])
			assert.NotNil(t, reqBody["endTime"])
			assert.Equal(t, float64(25), reqBody["completionTokens"])
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		endTime := time.Now()
		completionTokens := 25
		update := &domain.GenerationUpdate{
			Output: map[string]string{
				"response": "Hello! How can I help you?",
			},
			EndTime:          &endTime,
			CompletionTokens: &completionTokens,
		}
		
		// Act
		err := monitor.UpdateGeneration(ctx, generationID, update)
		
		// Assert
		assert.NoError(t, err)
	})
}

func TestLangfuseMonitor_ScoreGeneration(t *testing.T) {
	t.Run("successful score creation", func(t *testing.T) {
		generationID := uuid.New().String()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/public/scores", r.URL.Path)
			
			var reqBody map[string]any
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, generationID, reqBody["observationId"])
			assert.Equal(t, "quality", reqBody["name"])
			assert.Equal(t, float64(0.95), reqBody["value"])
			assert.Equal(t, "numeric", reqBody["dataType"])
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		score := &domain.Score{
			ID:        uuid.New().String(),
			Name:      "quality",
			Value:     0.95,
			DataType:  domain.ScoreDataTypeNumeric,
			Comment:   "High quality response",
			Timestamp: time.Now(),
		}
		
		// Act
		err := monitor.ScoreGeneration(ctx, generationID, score)
		
		// Assert
		assert.NoError(t, err)
	})
}

func TestLangfuseMonitor_GetTraceMetrics(t *testing.T) {
	t.Run("successful metrics retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/public/metrics/traces", r.URL.Path)
			
			// Check query parameters
			query := r.URL.Query()
			assert.NotEmpty(t, query.Get("fromTimestamp"))
			assert.NotEmpty(t, query.Get("toTimestamp"))
			assert.Equal(t, "user-123", query.Get("userId"))
			
			response := map[string]any{
				"totalTraces": 150,
				"totalGenerations": 450,
				"totalTokens": 125000,
				"averageLatencyMs": 1250,
				"successRate": 0.98,
				"errorRate": 0.02,
				"tokensPerTrace": 833.33,
				"costEstimate": 2.50,
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		filter := &domain.MetricsFilter{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now(),
			UserID:    "user-123",
		}
		
		// Act
		metrics, err := monitor.GetTraceMetrics(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, 150, metrics.TotalTraces)
		assert.Equal(t, 450, metrics.TotalGenerations)
		assert.Equal(t, 125000, metrics.TotalTokens)
		assert.Equal(t, 1250*time.Millisecond, metrics.AverageLatency)
		assert.Equal(t, 0.98, metrics.SuccessRate)
		assert.Equal(t, 2.50, metrics.CostEstimate)
	})
}

func TestLangfuseMonitor_CreateDataset(t *testing.T) {
	t.Run("successful dataset creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/public/datasets", r.URL.Path)
			
			var reqBody map[string]any
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, "test-dataset", reqBody["name"])
			assert.Equal(t, "Dataset for testing", reqBody["description"])
			
			response := map[string]any{
				"id": reqBody["id"],
				"success": true,
			}
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		dataset := &domain.Dataset{
			ID:          uuid.New().String(),
			Name:        "test-dataset",
			Description: "Dataset for testing",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Act
		err := monitor.CreateDataset(ctx, dataset)
		
		// Assert
		assert.NoError(t, err)
	})
}

func TestLangfuseMonitor_AddToDataset(t *testing.T) {
	t.Run("successful dataset item addition", func(t *testing.T) {
		datasetID := uuid.New().String()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/public/dataset-items", r.URL.Path)
			
			var reqBody map[string]any
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, datasetID, reqBody["datasetId"])
			assert.NotNil(t, reqBody["input"])
			assert.NotNil(t, reqBody["expectedOutput"])
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		item := &domain.DatasetItem{
			ID: uuid.New().String(),
			Input: map[string]string{
				"question": "What is the capital of France?",
			},
			ExpectedOutput: map[string]string{
				"answer": "Paris",
			},
			CreatedAt: time.Now(),
		}
		
		// Act
		err := monitor.AddToDataset(ctx, datasetID, item)
		
		// Assert
		assert.NoError(t, err)
	})
}

func TestLangfuseMonitor_IsHealthy(t *testing.T) {
	t.Run("healthy service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/public/health", r.URL.Path)
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		// Act
		healthy := monitor.IsHealthy(ctx)
		
		// Assert
		assert.True(t, healthy)
	})
	
	t.Run("unhealthy service", func(t *testing.T) {
		// Use invalid URL to simulate connection failure
		monitor := NewLangfuseMonitor("http://invalid-host:3000", "pk_test", "sk_test", 100*time.Millisecond)
		ctx := context.Background()
		
		// Act
		healthy := monitor.IsHealthy(ctx)
		
		// Assert
		assert.False(t, healthy)
	})
}

func TestLangfuseMonitor_Flush(t *testing.T) {
	t.Run("successful flush", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Flush endpoint would typically accept batched events
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/public/ingestion", r.URL.Path)
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		}))
		defer server.Close()
		
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		// Act
		err := monitor.Flush(ctx)
		
		// Assert
		assert.NoError(t, err)
	})
}

func TestLangfuseMonitor_BatchingBehavior(t *testing.T) {
	t.Run("events are batched before sending", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			
			if r.URL.Path == "/api/public/ingestion" {
				// Batch endpoint
				var reqBody map[string]any
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				
				// Should contain multiple events
				events, ok := reqBody["batch"].([]any)
				assert.True(t, ok)
				assert.Greater(t, len(events), 1) // Should batch multiple events
			}
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		}))
		defer server.Close()
		
		// Create monitor with batching enabled
		monitor := NewLangfuseMonitor(server.URL, "pk_test", "sk_test", 30*time.Second)
		ctx := context.Background()
		
		// Create multiple traces quickly
		for i := 0; i < 5; i++ {
			trace := &domain.Trace{
				ID:        uuid.New().String(),
				Name:      "batch-test-trace",
				UserID:    "user-123",
				Timestamp: time.Now(),
			}
			err := monitor.CreateTrace(ctx, trace)
			assert.NoError(t, err)
		}
		
		// Flush to ensure all events are sent
		err := monitor.Flush(ctx)
		assert.NoError(t, err)
		
		// The implementation sends both immediate and batched requests
		// So we expect more than 5 requests (5 immediate + batch requests)
		assert.GreaterOrEqual(t, requestCount, 5)
	})
}