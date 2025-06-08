package aiops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/aiops"
	"github.com/stretchr/testify/assert"
)

func TestOllamaProvider_Chat(t *testing.T) {
	t.Run("successful chat request", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/chat", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			// Mock response
			response := `{
				"model": "llama2:7b",
				"created_at": "2023-01-01T00:00:00Z",
				"message": {
					"role": "assistant",
					"content": "I can help you with Kubernetes troubleshooting. What specific issue are you experiencing?"
				},
				"done": true,
				"total_duration": 1500000000,
				"load_duration": 100000000,
				"prompt_eval_duration": 200000000,
				"eval_duration": 300000000,
				"eval_count": 25
			}`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		// Create Ollama provider
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Prepare request
		req := &aiops.ChatRequest{
			Model: "llama2:7b",
			Messages: []aiops.ChatMessage{
				{Role: "user", Content: "Help me troubleshoot Kubernetes"},
			},
		}
		
		// Act
		response, err := provider.Chat(ctx, req)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "llama2:7b", response.Model)
		assert.Equal(t, "assistant", response.Message.Role)
		assert.Contains(t, response.Message.Content, "Kubernetes troubleshooting")
		assert.True(t, response.Done)
		assert.NotNil(t, response.Usage)
		assert.Equal(t, 25, response.Usage.CompletionTokens)
	})
	
	t.Run("server error", func(t *testing.T) {
		// Create mock server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		req := &aiops.ChatRequest{
			Model: "llama2:7b",
			Messages: []aiops.ChatMessage{
				{Role: "user", Content: "Test"},
			},
		}
		
		// Act
		response, err := provider.Chat(ctx, req)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "500")
	})
	
	t.Run("network timeout", func(t *testing.T) {
		// Create mock server that delays response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		
		// Set short timeout
		provider := NewOllamaProvider(server.URL, 100*time.Millisecond, nil)
		ctx := context.Background()
		
		req := &aiops.ChatRequest{
			Model: "llama2:7b",
			Messages: []aiops.ChatMessage{
				{Role: "user", Content: "Test"},
			},
		}
		
		// Act
		response, err := provider.Chat(ctx, req)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "exceeded")
	})
}

func TestOllamaProvider_ListModels(t *testing.T) {
	t.Run("successful model listing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/tags", r.URL.Path)
			
			response := `{
				"models": [
					{
						"name": "llama2:7b",
						"modified_at": "2023-01-01T00:00:00Z",
						"size": 3800000000,
						"digest": "sha256:abc123",
						"details": {
							"format": "gguf",
							"family": "llama",
							"parameter": "7B",
							"quantization": "Q4_0"
						}
					},
					{
						"name": "codellama:13b",
						"modified_at": "2023-01-02T00:00:00Z",
						"size": 7300000000,
						"digest": "sha256:def456",
						"details": {
							"format": "gguf",
							"family": "llama",
							"parameter": "13B",
							"quantization": "Q4_0"
						}
					}
				]
			}`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Act
		models, err := provider.ListModels(ctx)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, models, 2)
		assert.Equal(t, "llama2:7b", models[0].Name)
		assert.Equal(t, int64(3800000000), models[0].Size)
		assert.Equal(t, "codellama:13b", models[1].Name)
		assert.Equal(t, int64(7300000000), models[1].Size)
		assert.Equal(t, aiops.ModelStatusAvailable, models[0].Status)
	})
}

func TestOllamaProvider_PullModel(t *testing.T) {
	t.Run("successful model pull", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/pull", r.URL.Path)
			
			// Simulate streaming response
			responses := []string{
				`{"status":"downloading","digest":"sha256:abc123","total":3800000000,"completed":1000000000}`,
				`{"status":"downloading","digest":"sha256:abc123","total":3800000000,"completed":2000000000}`,
				`{"status":"downloading","digest":"sha256:abc123","total":3800000000,"completed":3800000000}`,
				`{"status":"success"}`,
			}
			
			w.Header().Set("Content-Type", "application/x-ndjson")
			w.WriteHeader(http.StatusOK)
			
			for _, resp := range responses {
				w.Write([]byte(resp + "\n"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(10 * time.Millisecond)
			}
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Act
		err := provider.PullModel(ctx, "llama2:7b")
		
		// Assert
		assert.NoError(t, err)
	})
	
	t.Run("model pull failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"model not found"}`))
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Act
		err := provider.PullModel(ctx, "nonexistent:model")
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestOllamaProvider_StreamChat(t *testing.T) {
	t.Run("successful streaming chat", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/chat", r.URL.Path)
			
			// Simulate streaming response
			responses := []string{
				`{"model":"llama2:7b","created_at":"2023-01-01T00:00:00Z","message":{"role":"assistant","content":"I"},"done":false}`,
				`{"model":"llama2:7b","created_at":"2023-01-01T00:00:00Z","message":{"role":"assistant","content":" can"},"done":false}`,
				`{"model":"llama2:7b","created_at":"2023-01-01T00:00:00Z","message":{"role":"assistant","content":" help"},"done":false}`,
				`{"model":"llama2:7b","created_at":"2023-01-01T00:00:00Z","message":{"role":"assistant","content":" you"},"done":true,"eval_count":25}`,
			}
			
			w.Header().Set("Content-Type", "application/x-ndjson")
			w.WriteHeader(http.StatusOK)
			
			for _, resp := range responses {
				w.Write([]byte(resp + "\n"))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(10 * time.Millisecond)
			}
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		req := &aiops.ChatRequest{
			Model: "llama2:7b",
			Messages: []aiops.ChatMessage{
				{Role: "user", Content: "Help me"},
			},
			Stream: true,
		}
		
		// Act
		responseChan, err := provider.StreamChat(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, responseChan)
		
		// Collect responses
		var responses []*aiops.ChatStreamResponse
		for response := range responseChan {
			responses = append(responses, response)
		}
		
		// Assert
		assert.Len(t, responses, 4)
		assert.False(t, responses[0].Done)
		assert.False(t, responses[1].Done)
		assert.False(t, responses[2].Done)
		assert.True(t, responses[3].Done)
		
		// Verify content progression
		assert.Equal(t, "I", responses[0].Message.Content)
		assert.Equal(t, " can", responses[1].Message.Content)
		assert.Equal(t, " help", responses[2].Message.Content)
		assert.Equal(t, " you", responses[3].Message.Content)
	})
}

func TestOllamaProvider_IsHealthy(t *testing.T) {
	t.Run("healthy service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/", r.URL.Path)
			
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ollama is running"))
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Act
		healthy := provider.IsHealthy(ctx)
		
		// Assert
		assert.True(t, healthy)
	})
	
	t.Run("unhealthy service", func(t *testing.T) {
		// Use invalid URL to simulate connection failure
		provider := NewOllamaProvider("http://invalid-host:11434", 100*time.Millisecond, nil)
		ctx := context.Background()
		
		// Act
		healthy := provider.IsHealthy(ctx)
		
		// Assert
		assert.False(t, healthy)
	})
}

func TestOllamaProvider_GetModelInfo(t *testing.T) {
	t.Run("successful model info retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/show", r.URL.Path)
			
			response := `{
				"license": "MIT",
				"modelfile": "FROM llama2:7b\nSYSTEM You are a helpful assistant.",
				"parameters": "num_ctx 4096\nstop [\"<|start_header_id|>\",\"<|end_header_id|>\",\"<|eot_id|>\"]",
				"template": "{{ if .System }}<|start_header_id|>system<|end_header_id|>\n\n{{ .System }}<|eot_id|>{{ end }}{{ if .Prompt }}<|start_header_id|>user<|end_header_id|>\n\n{{ .Prompt }}<|eot_id|>{{ end }}<|start_header_id|>assistant<|end_header_id|>\n\n",
				"details": {
					"format": "gguf",
					"family": "llama",
					"families": ["llama"],
					"parameter_size": "7B",
					"quantization_level": "Q4_0"
				}
			}`
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Act
		info, err := provider.GetModelInfo(ctx, "llama2:7b")
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "llama2:7b", info.Name)
		assert.Equal(t, aiops.ModelStatusAvailable, info.Status)
		assert.NotNil(t, info.Details)
		assert.Equal(t, "gguf", info.Details.Format)
		assert.Equal(t, "llama", info.Details.Family)
		assert.Equal(t, "7B", info.Details.Parameter)
		assert.Equal(t, "Q4_0", info.Details.Quantization)
	})
	
	t.Run("model not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"model not found"}`))
		}))
		defer server.Close()
		
		provider := NewOllamaProvider(server.URL, 30*time.Second, nil)
		ctx := context.Background()
		
		// Act
		info, err := provider.GetModelInfo(ctx, "nonexistent:model")
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "404")
	})
}