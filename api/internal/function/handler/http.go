package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
)

// Handler handles HTTP requests for function management
type Handler struct {
	service domain.Service
	logger  *slog.Logger
}

// NewHandler creates a new function handler
func NewHandler(service domain.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// ListFunctions handles GET /workspaces/{workspaceID}/projects/{projectID}/functions
func (h *Handler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	projectID := vars["projectID"]

	functions, err := h.service.ListFunctions(r.Context(), workspaceID, projectID)
	if err != nil {
		h.logger.Error("failed to list functions", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(functions)
}

// CreateFunction handles POST /workspaces/{workspaceID}/projects/{projectID}/functions
func (h *Handler) CreateFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	projectID := vars["projectID"]

	var spec domain.FunctionSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fn, err := h.service.CreateFunction(r.Context(), workspaceID, projectID, &spec)
	if err != nil {
		h.logger.Error("failed to create function", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fn)
}

// GetFunction handles GET /workspaces/{workspaceID}/functions/{functionID}
func (h *Handler) GetFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	fn, err := h.service.GetFunction(r.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("failed to get function", "error", err)
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fn)
}

// UpdateFunction handles PUT /workspaces/{workspaceID}/functions/{functionID}
func (h *Handler) UpdateFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	var spec domain.FunctionSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fn, err := h.service.UpdateFunction(r.Context(), workspaceID, functionID, &spec)
	if err != nil {
		h.logger.Error("failed to update function", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fn)
}

// DeleteFunction handles DELETE /workspaces/{workspaceID}/functions/{functionID}
func (h *Handler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	if err := h.service.DeleteFunction(r.Context(), workspaceID, functionID); err != nil {
		h.logger.Error("failed to delete function", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListVersions handles GET /workspaces/{workspaceID}/functions/{functionID}/versions
func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	versions, err := h.service.ListVersions(r.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("failed to list versions", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

// DeployVersion handles POST /workspaces/{workspaceID}/functions/{functionID}/versions
func (h *Handler) DeployVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	var version domain.FunctionVersionDef
	if err := json.NewDecoder(r.Body).Decode(&version); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	deployedVersion, err := h.service.DeployVersion(r.Context(), workspaceID, functionID, &version)
	if err != nil {
		h.logger.Error("failed to deploy version", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployedVersion)
}

// SetActiveVersion handles PUT /workspaces/{workspaceID}/functions/{functionID}/versions/{versionID}/activate
func (h *Handler) SetActiveVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]
	versionID := vars["versionID"]

	if err := h.service.SetActiveVersion(r.Context(), workspaceID, functionID, versionID); err != nil {
		h.logger.Error("failed to set active version", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTriggers handles GET /workspaces/{workspaceID}/functions/{functionID}/triggers
func (h *Handler) ListTriggers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	triggers, err := h.service.ListTriggers(r.Context(), workspaceID, functionID)
	if err != nil {
		h.logger.Error("failed to list triggers", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggers)
}

// CreateTrigger handles POST /workspaces/{workspaceID}/functions/{functionID}/triggers
func (h *Handler) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	var trigger domain.FunctionTrigger
	if err := json.NewDecoder(r.Body).Decode(&trigger); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdTrigger, err := h.service.CreateTrigger(r.Context(), workspaceID, functionID, &trigger)
	if err != nil {
		h.logger.Error("failed to create trigger", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTrigger)
}

// InvokeFunction handles POST /workspaces/{workspaceID}/functions/{functionID}/invoke
func (h *Handler) InvokeFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	// Check if async
	async := r.URL.Query().Get("async") == "true"

	var request domain.InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Allow empty body
		request = domain.InvokeRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header,
		}
	}

	if async {
		invocationID, err := h.service.InvokeFunctionAsync(r.Context(), workspaceID, functionID, &request)
		if err != nil {
			h.logger.Error("failed to invoke function async", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"invocationID": invocationID,
		})
	} else {
		response, err := h.service.InvokeFunction(r.Context(), workspaceID, functionID, &request)
		if err != nil {
			h.logger.Error("failed to invoke function", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Copy response headers
		for k, v := range response.Headers {
			w.Header()[k] = v
		}
		w.WriteHeader(response.StatusCode)
		w.Write(response.Body)
	}
}

// GetInvocationStatus handles GET /workspaces/{workspaceID}/invocations/{invocationID}
func (h *Handler) GetInvocationStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	invocationID := vars["invocationID"]

	status, err := h.service.GetInvocationStatus(r.Context(), workspaceID, invocationID)
	if err != nil {
		h.logger.Error("failed to get invocation status", "error", err)
		http.Error(w, "Invocation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetFunctionLogs handles GET /workspaces/{workspaceID}/functions/{functionID}/logs
func (h *Handler) GetFunctionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	opts := &domain.LogOptions{
		Follow: r.URL.Query().Get("follow") == "true",
		Limit:  100,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			opts.Limit = limit
		}
	}

	logs, err := h.service.GetFunctionLogs(r.Context(), workspaceID, functionID, opts)
	if err != nil {
		h.logger.Error("failed to get function logs", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetFunctionMetrics handles GET /workspaces/{workspaceID}/functions/{functionID}/metrics
func (h *Handler) GetFunctionMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]
	functionID := vars["functionID"]

	// Default to last hour
	now := time.Now()
	opts := &domain.MetricOptions{
		StartTime:  now.Add(-1 * time.Hour),
		EndTime:    now,
		Resolution: "5m",
	}

	if resolution := r.URL.Query().Get("resolution"); resolution != "" {
		opts.Resolution = resolution
	}

	metrics, err := h.service.GetFunctionMetrics(r.Context(), workspaceID, functionID, opts)
	if err != nil {
		h.logger.Error("failed to get function metrics", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetProviderCapabilities handles GET /workspaces/{workspaceID}/functions/capabilities
func (h *Handler) GetProviderCapabilities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceID"]

	capabilities, err := h.service.GetProviderCapabilities(r.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to get provider capabilities", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(capabilities)
}