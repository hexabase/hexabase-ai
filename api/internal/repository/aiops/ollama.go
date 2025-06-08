package aiops

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/aiops"
)

// OllamaProvider implements the LLMService interface for Ollama
type OllamaProvider struct {
	baseURL    string
	timeout    time.Duration
	headers    map[string]string
	httpClient *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string, timeout time.Duration, headers map[string]string) aiops.LLMService {
	if headers == nil {
		headers = make(map[string]string)
	}
	
	return &OllamaProvider{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		timeout: timeout,
		headers: headers,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat sends a chat request to Ollama
func (p *OllamaProvider) Chat(ctx context.Context, req *aiops.ChatRequest) (*aiops.ChatResponse, error) {
	// Convert to Ollama request format
	ollamaReq := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
	}
	
	if req.Temperature != nil {
		if ollamaReq["options"] == nil {
			ollamaReq["options"] = make(map[string]any)
		}
		ollamaReq["options"].(map[string]any)["temperature"] = *req.Temperature
	}
	
	if req.MaxTokens != nil {
		if ollamaReq["options"] == nil {
			ollamaReq["options"] = make(map[string]any)
		}
		ollamaReq["options"].(map[string]any)["num_predict"] = *req.MaxTokens
	}
	
	// Add additional options
	for k, v := range req.Options {
		if ollamaReq["options"] == nil {
			ollamaReq["options"] = make(map[string]any)
		}
		ollamaReq["options"].(map[string]any)[k] = v
	}
	
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var ollamaResp struct {
		Model     string `json:"model"`
		CreatedAt string `json:"created_at"`
		Message   struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done             bool `json:"done"`
		TotalDuration    int64 `json:"total_duration"`
		LoadDuration     int64 `json:"load_duration"`
		PromptEvalCount  int   `json:"prompt_eval_count"`
		PromptEvalDuration int64 `json:"prompt_eval_duration"`
		EvalCount        int   `json:"eval_count"`
		EvalDuration     int64 `json:"eval_duration"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert created_at to time
	createdAt, err := time.Parse(time.RFC3339, ollamaResp.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}
	
	// Convert to domain response
	response := &aiops.ChatResponse{
		Model: ollamaResp.Model,
		Message: aiops.ChatMessage{
			Role:    ollamaResp.Message.Role,
			Content: ollamaResp.Message.Content,
		},
		Done:      ollamaResp.Done,
		CreatedAt: createdAt,
	}
	
	// Calculate usage statistics from Ollama response
	if ollamaResp.EvalCount > 0 || ollamaResp.PromptEvalCount > 0 {
		response.Usage = &aiops.UsageStats{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		}
	}
	
	return response, nil
}

// StreamChat sends a streaming chat request to Ollama
func (p *OllamaProvider) StreamChat(ctx context.Context, req *aiops.ChatRequest) (<-chan *aiops.ChatStreamResponse, error) {
	// Convert to Ollama request format
	ollamaReq := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}
	
	if req.Temperature != nil {
		if ollamaReq["options"] == nil {
			ollamaReq["options"] = make(map[string]any)
		}
		ollamaReq["options"].(map[string]any)["temperature"] = *req.Temperature
	}
	
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	// Create response channel
	responseChan := make(chan *aiops.ChatStreamResponse, 10)
	
	go func() {
		defer close(responseChan)
		defer resp.Body.Close()
		
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			
			var ollamaResp struct {
				Model     string `json:"model"`
				CreatedAt string `json:"created_at"`
				Message   struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Done      bool `json:"done"`
				EvalCount int  `json:"eval_count"`
				Error     string `json:"error,omitempty"`
			}
			
			err := json.Unmarshal([]byte(line), &ollamaResp)
			if err != nil {
				responseChan <- &aiops.ChatStreamResponse{
					Error: fmt.Sprintf("failed to decode response: %v", err),
				}
				return
			}
			
			if ollamaResp.Error != "" {
				responseChan <- &aiops.ChatStreamResponse{
					Error: ollamaResp.Error,
				}
				return
			}
			
			// Convert created_at to time
			createdAt, err := time.Parse(time.RFC3339, ollamaResp.CreatedAt)
			if err != nil {
				createdAt = time.Now()
			}
			
			// Send response chunk
			responseChan <- &aiops.ChatStreamResponse{
				Model: ollamaResp.Model,
				Message: aiops.ChatMessage{
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				Done:      ollamaResp.Done,
				CreatedAt: createdAt,
			}
			
			// Exit on done
			if ollamaResp.Done {
				break
			}
		}
		
		if err := scanner.Err(); err != nil {
			responseChan <- &aiops.ChatStreamResponse{
				Error: fmt.Sprintf("stream error: %v", err),
			}
		}
	}()
	
	return responseChan, nil
}

// ListModels lists available models from Ollama
func (p *OllamaProvider) ListModels(ctx context.Context) ([]*aiops.ModelInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var ollamaResp struct {
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Digest     string `json:"digest"`
			Details    struct {
				Format       string   `json:"format"`
				Family       string   `json:"family"`
				Families     []string `json:"families"`
				Parameter    string   `json:"parameter"`
				Quantization string   `json:"quantization"`
			} `json:"details"`
		} `json:"models"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	models := make([]*aiops.ModelInfo, 0, len(ollamaResp.Models))
	for _, model := range ollamaResp.Models {
		modifiedAt, err := time.Parse(time.RFC3339, model.ModifiedAt)
		if err != nil {
			modifiedAt = time.Now()
		}
		
		modelInfo := &aiops.ModelInfo{
			Name:       model.Name,
			ModifiedAt: modifiedAt,
			Size:       model.Size,
			Digest:     model.Digest,
			Status:     aiops.ModelStatusAvailable,
			Details: &aiops.ModelDetails{
				Format:       model.Details.Format,
				Family:       model.Details.Family,
				Families:     model.Details.Families,
				Parameter:    model.Details.Parameter,
				Quantization: model.Details.Quantization,
			},
		}
		
		models = append(models, modelInfo)
	}
	
	return models, nil
}

// PullModel pulls a model from Ollama registry
func (p *OllamaProvider) PullModel(ctx context.Context, modelName string) error {
	reqBody, err := json.Marshal(map[string]string{
		"name": modelName,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	// Read streaming response to completion
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		var pullResp struct {
			Status    string `json:"status"`
			Error     string `json:"error,omitempty"`
			Total     int64  `json:"total,omitempty"`
			Completed int64  `json:"completed,omitempty"`
		}
		
		err := json.Unmarshal([]byte(line), &pullResp)
		if err != nil {
			continue // Skip malformed lines
		}
		
		if pullResp.Error != "" {
			return fmt.Errorf("model pull error: %s", pullResp.Error)
		}
		
		// Check for success status
		if pullResp.Status == "success" {
			return nil
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream error: %w", err)
	}
	
	return nil
}

// DeleteModel deletes a model from Ollama
func (p *OllamaProvider) DeleteModel(ctx context.Context, modelName string) error {
	reqBody, err := json.Marshal(map[string]string{
		"name": modelName,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", p.baseURL+"/api/delete", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// IsHealthy checks if Ollama service is healthy
func (p *OllamaProvider) IsHealthy(ctx context.Context) bool {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/", nil)
	if err != nil {
		return false
	}
	
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// GetModelInfo gets information about a specific model
func (p *OllamaProvider) GetModelInfo(ctx context.Context, modelName string) (*aiops.ModelInfo, error) {
	reqBody, err := json.Marshal(map[string]string{
		"name": modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/show", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}
	
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var ollamaResp struct {
		License    string `json:"license"`
		Modelfile  string `json:"modelfile"`
		Parameters string `json:"parameters"`
		Template   string `json:"template"`
		Details    struct {
			Format           string   `json:"format"`
			Family           string   `json:"family"`
			Families         []string `json:"families"`
			ParameterSize    string   `json:"parameter_size"`
			QuantizationLevel string  `json:"quantization_level"`
		} `json:"details"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	modelInfo := &aiops.ModelInfo{
		Name:   modelName,
		Status: aiops.ModelStatusAvailable,
		Details: &aiops.ModelDetails{
			Format:       ollamaResp.Details.Format,
			Family:       ollamaResp.Details.Family,
			Families:     ollamaResp.Details.Families,
			Parameter:    ollamaResp.Details.ParameterSize,
			Quantization: ollamaResp.Details.QuantizationLevel,
		},
		Metadata: map[string]any{
			"license":    ollamaResp.License,
			"modelfile":  ollamaResp.Modelfile,
			"parameters": ollamaResp.Parameters,
			"template":   ollamaResp.Template,
		},
	}
	
	return modelInfo, nil
}