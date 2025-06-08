package aiops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/aiops"
)

// LangfuseMonitor implements the LLMOpsMonitor interface for Langfuse
type LangfuseMonitor struct {
	baseURL    string
	publicKey  string
	secretKey  string
	httpClient *http.Client
	
	// Batching support
	batchMutex   sync.Mutex
	batchQueue   []batchEvent
	batchSize    int
	flushInterval time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// batchEvent represents an event to be batched
type batchEvent struct {
	Type      string    `json:"type"`
	Body      any       `json:"body"`
	Timestamp time.Time `json:"timestamp"`
}

// NewLangfuseMonitor creates a new Langfuse monitor
func NewLangfuseMonitor(baseURL, publicKey, secretKey string, timeout time.Duration) aiops.LLMOpsMonitor {
	monitor := &LangfuseMonitor{
		baseURL:      baseURL,
		publicKey:    publicKey,
		secretKey:    secretKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		batchQueue:    make([]batchEvent, 0, 100),
		batchSize:     50,
		flushInterval: 5 * time.Second,
		stopChan:      make(chan struct{}),
	}
	
	// Start background flushing
	monitor.wg.Add(1)
	go monitor.backgroundFlush()
	
	return monitor
}

// CreateTrace creates a new trace in Langfuse
func (m *LangfuseMonitor) CreateTrace(ctx context.Context, trace *aiops.Trace) error {
	payload := map[string]any{
		"id":        trace.ID,
		"name":      trace.Name,
		"userId":    trace.UserID,
		"sessionId": trace.SessionID,
		"timestamp": trace.Timestamp.Format(time.RFC3339),
		"tags":      trace.Tags,
		"metadata":  trace.Metadata,
		"input":     trace.Input,
		"output":    trace.Output,
		"release":   trace.Release,
		"version":   trace.Version,
	}
	
	// Add to batch queue
	m.addToBatch("trace-create", payload)
	
	// For immediate feedback, also send directly
	return m.sendRequest(ctx, "POST", "/api/public/traces", payload)
}

// CreateGeneration creates a new generation in Langfuse
func (m *LangfuseMonitor) CreateGeneration(ctx context.Context, generation *aiops.Generation) error {
	payload := map[string]any{
		"id":               generation.ID,
		"traceId":          generation.TraceID,
		"parentObservationId": generation.ParentID,
		"name":             generation.Name,
		"model":            generation.Model,
		"modelParameters":  generation.ModelParameters,
		"input":            generation.Input,
		"output":           generation.Output,
		"startTime":        generation.StartTime.Format(time.RFC3339),
		"completionTokens": generation.CompletionTokens,
		"promptTokens":     generation.PromptTokens,
		"totalTokens":      generation.TotalTokens,
		"statusMessage":    generation.StatusMessage,
		"level":            generation.Level,
		"metadata":         generation.Metadata,
	}
	
	if generation.EndTime != nil {
		payload["endTime"] = generation.EndTime.Format(time.RFC3339)
	}
	
	// Add to batch queue
	m.addToBatch("generation-create", payload)
	
	// For immediate feedback, also send directly
	return m.sendRequest(ctx, "POST", "/api/public/generations", payload)
}

// CreateSpan creates a new span in Langfuse
func (m *LangfuseMonitor) CreateSpan(ctx context.Context, span *aiops.Span) error {
	payload := map[string]any{
		"id":               span.ID,
		"traceId":          span.TraceID,
		"parentObservationId": span.ParentID,
		"name":             span.Name,
		"startTime":        span.StartTime.Format(time.RFC3339),
		"input":            span.Input,
		"output":           span.Output,
		"statusMessage":    span.StatusMessage,
		"level":            span.Level,
		"metadata":         span.Metadata,
	}
	
	if span.EndTime != nil {
		payload["endTime"] = span.EndTime.Format(time.RFC3339)
	}
	
	return m.sendRequest(ctx, "POST", "/api/public/spans", payload)
}

// UpdateGeneration updates an existing generation
func (m *LangfuseMonitor) UpdateGeneration(ctx context.Context, generationID string, updates *aiops.GenerationUpdate) error {
	payload := make(map[string]any)
	
	if updates.Output != nil {
		payload["output"] = updates.Output
	}
	if updates.EndTime != nil {
		payload["endTime"] = updates.EndTime.Format(time.RFC3339)
	}
	if updates.CompletionTokens != nil {
		payload["completionTokens"] = *updates.CompletionTokens
	}
	if updates.PromptTokens != nil {
		payload["promptTokens"] = *updates.PromptTokens
	}
	if updates.TotalTokens != nil {
		payload["totalTokens"] = *updates.TotalTokens
	}
	if updates.StatusMessage != nil {
		payload["statusMessage"] = *updates.StatusMessage
	}
	if updates.Metadata != nil {
		payload["metadata"] = updates.Metadata
	}
	
	return m.sendRequest(ctx, "PATCH", fmt.Sprintf("/api/public/generations/%s", generationID), payload)
}

// ScoreGeneration creates a score for a generation
func (m *LangfuseMonitor) ScoreGeneration(ctx context.Context, generationID string, score *aiops.Score) error {
	payload := map[string]any{
		"id":             score.ID,
		"name":           score.Name,
		"value":          score.Value,
		"dataType":       string(score.DataType),
		"observationId":  generationID,
		"timestamp":      score.Timestamp.Format(time.RFC3339),
	}
	
	if score.StringValue != "" {
		payload["stringValue"] = score.StringValue
	}
	if score.Comment != "" {
		payload["comment"] = score.Comment
	}
	if score.ObserverID != "" {
		payload["observerId"] = score.ObserverID
	}
	
	return m.sendRequest(ctx, "POST", "/api/public/scores", payload)
}

// ScoreTrace creates a score for a trace
func (m *LangfuseMonitor) ScoreTrace(ctx context.Context, traceID string, score *aiops.Score) error {
	payload := map[string]any{
		"id":        score.ID,
		"name":      score.Name,
		"value":     score.Value,
		"dataType":  string(score.DataType),
		"traceId":   traceID,
		"timestamp": score.Timestamp.Format(time.RFC3339),
	}
	
	if score.StringValue != "" {
		payload["stringValue"] = score.StringValue
	}
	if score.Comment != "" {
		payload["comment"] = score.Comment
	}
	if score.ObserverID != "" {
		payload["observerId"] = score.ObserverID
	}
	
	return m.sendRequest(ctx, "POST", "/api/public/scores", payload)
}

// GetTraceMetrics retrieves trace metrics
func (m *LangfuseMonitor) GetTraceMetrics(ctx context.Context, filter *aiops.MetricsFilter) (*aiops.TraceMetrics, error) {
	params := url.Values{}
	params.Set("fromTimestamp", filter.StartTime.Format(time.RFC3339))
	params.Set("toTimestamp", filter.EndTime.Format(time.RFC3339))
	
	if filter.UserID != "" {
		params.Set("userId", filter.UserID)
	}
	if filter.SessionID != "" {
		params.Set("sessionId", filter.SessionID)
	}
	if filter.Model != "" {
		params.Set("model", filter.Model)
	}
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			params.Add("tags", tag)
		}
	}
	if filter.Release != "" {
		params.Set("release", filter.Release)
	}
	
	var response struct {
		TotalTraces      int                `json:"totalTraces"`
		TotalGenerations int                `json:"totalGenerations"`
		TotalTokens      int                `json:"totalTokens"`
		AverageLatencyMs int                `json:"averageLatencyMs"`
		SuccessRate      float64            `json:"successRate"`
		ErrorRate        float64            `json:"errorRate"`
		TokensPerTrace   float64            `json:"tokensPerTrace"`
		CostEstimate     float64            `json:"costEstimate"`
		ScoreDistribution map[string]float64 `json:"scoreDistribution"`
	}
	
	err := m.getRequest(ctx, "/api/public/metrics/traces", params, &response)
	if err != nil {
		return nil, err
	}
	
	return &aiops.TraceMetrics{
		TotalTraces:      response.TotalTraces,
		TotalGenerations: response.TotalGenerations,
		TotalTokens:      response.TotalTokens,
		AverageLatency:   time.Duration(response.AverageLatencyMs) * time.Millisecond,
		SuccessRate:      response.SuccessRate,
		ErrorRate:        response.ErrorRate,
		TokensPerTrace:   response.TokensPerTrace,
		CostEstimate:     response.CostEstimate,
		ScoreDistribution: response.ScoreDistribution,
	}, nil
}

// GetModelMetrics retrieves model-specific metrics
func (m *LangfuseMonitor) GetModelMetrics(ctx context.Context, filter *aiops.MetricsFilter) (*aiops.LLMModelMetrics, error) {
	params := url.Values{}
	params.Set("fromTimestamp", filter.StartTime.Format(time.RFC3339))
	params.Set("toTimestamp", filter.EndTime.Format(time.RFC3339))
	
	if filter.Model != "" {
		params.Set("model", filter.Model)
	}
	
	var response struct {
		ModelName        string             `json:"modelName"`
		TotalGenerations int                `json:"totalGenerations"`
		TotalTokens      int                `json:"totalTokens"`
		AverageLatencyMs int                `json:"averageLatencyMs"`
		TokensPerSecond  float64            `json:"tokensPerSecond"`
		CostPerToken     float64            `json:"costPerToken"`
		ErrorRate        float64            `json:"errorRate"`
		ModelVersions    []string           `json:"modelVersions"`
		ScoreStats       map[string]float64 `json:"scoreStats"`
	}
	
	err := m.getRequest(ctx, "/api/public/metrics/models", params, &response)
	if err != nil {
		return nil, err
	}
	
	return &aiops.LLMModelMetrics{
		ModelName:        response.ModelName,
		TotalGenerations: response.TotalGenerations,
		TotalTokens:      response.TotalTokens,
		AverageLatency:   time.Duration(response.AverageLatencyMs) * time.Millisecond,
		TokensPerSecond:  response.TokensPerSecond,
		CostPerToken:     response.CostPerToken,
		ErrorRate:        response.ErrorRate,
		ModelVersions:    response.ModelVersions,
		ScoreStats:       response.ScoreStats,
	}, nil
}

// GetUserMetrics retrieves user-specific metrics
func (m *LangfuseMonitor) GetUserMetrics(ctx context.Context, filter *aiops.MetricsFilter) (*aiops.UserMetrics, error) {
	params := url.Values{}
	params.Set("fromTimestamp", filter.StartTime.Format(time.RFC3339))
	params.Set("toTimestamp", filter.EndTime.Format(time.RFC3339))
	
	if filter.UserID != "" {
		params.Set("userId", filter.UserID)
	}
	
	var response aiops.UserMetrics
	err := m.getRequest(ctx, "/api/public/metrics/users", params, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

// CreateDataset creates a new dataset
func (m *LangfuseMonitor) CreateDataset(ctx context.Context, dataset *aiops.Dataset) error {
	payload := map[string]any{
		"id":          dataset.ID,
		"name":        dataset.Name,
		"description": dataset.Description,
		"metadata":    dataset.Metadata,
	}
	
	return m.sendRequest(ctx, "POST", "/api/public/datasets", payload)
}

// AddToDataset adds an item to a dataset
func (m *LangfuseMonitor) AddToDataset(ctx context.Context, datasetID string, item *aiops.DatasetItem) error {
	payload := map[string]any{
		"id":             item.ID,
		"datasetId":      datasetID,
		"input":          item.Input,
		"expectedOutput": item.ExpectedOutput,
		"metadata":       item.Metadata,
	}
	
	if item.SourceTraceID != "" {
		payload["sourceTraceId"] = item.SourceTraceID
	}
	if item.SourceObservationID != "" {
		payload["sourceObservationId"] = item.SourceObservationID
	}
	
	return m.sendRequest(ctx, "POST", "/api/public/dataset-items", payload)
}

// GetDataset retrieves a dataset
func (m *LangfuseMonitor) GetDataset(ctx context.Context, datasetID string) (*aiops.Dataset, error) {
	var dataset aiops.Dataset
	err := m.getRequest(ctx, fmt.Sprintf("/api/public/datasets/%s", datasetID), nil, &dataset)
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

// IsHealthy checks if Langfuse service is healthy
func (m *LangfuseMonitor) IsHealthy(ctx context.Context) bool {
	var response map[string]any
	err := m.getRequest(ctx, "/api/public/health", nil, &response)
	return err == nil
}

// Flush flushes any pending events
func (m *LangfuseMonitor) Flush(ctx context.Context) error {
	m.batchMutex.Lock()
	events := make([]batchEvent, len(m.batchQueue))
	copy(events, m.batchQueue)
	m.batchQueue = m.batchQueue[:0]
	m.batchMutex.Unlock()
	
	if len(events) == 0 {
		return nil
	}
	
	payload := map[string]any{
		"batch": events,
	}
	
	return m.sendRequest(ctx, "POST", "/api/public/ingestion", payload)
}

// sendRequest sends an HTTP request to Langfuse
func (m *LangfuseMonitor) sendRequest(ctx context.Context, method, path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, m.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Langfuse-Public-Key", m.publicKey)
	req.Header.Set("X-Langfuse-Secret-Key", m.secretKey)
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("langfuse API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// getRequest sends a GET request to Langfuse
func (m *LangfuseMonitor) getRequest(ctx context.Context, path string, params url.Values, response any) error {
	reqURL := m.baseURL + path
	if params != nil && len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("X-Langfuse-Public-Key", m.publicKey)
	req.Header.Set("X-Langfuse-Secret-Key", m.secretKey)
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("langfuse API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	if response != nil {
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

// addToBatch adds an event to the batch queue
func (m *LangfuseMonitor) addToBatch(eventType string, body any) {
	m.batchMutex.Lock()
	defer m.batchMutex.Unlock()
	
	event := batchEvent{
		Type:      eventType,
		Body:      body,
		Timestamp: time.Now(),
	}
	
	m.batchQueue = append(m.batchQueue, event)
	
	// Flush if batch is full
	if len(m.batchQueue) >= m.batchSize {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			m.Flush(ctx)
		}()
	}
}

// backgroundFlush periodically flushes the batch queue
func (m *LangfuseMonitor) backgroundFlush() {
	defer m.wg.Done()
	ticker := time.NewTicker(m.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			m.Flush(ctx)
			cancel()
		case <-m.stopChan:
			// Final flush before stopping
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			m.Flush(ctx)
			cancel()
			return
		}
	}
}