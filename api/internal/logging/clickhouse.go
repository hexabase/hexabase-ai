package logging

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ClickHouseConfig holds ClickHouse connection configuration
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	TLS      bool
}

// ClickHouseLogger implements structured logging to ClickHouse
type ClickHouseLogger struct {
	conn   driver.Conn
	config *ClickHouseConfig
}

// LogEntry represents a structured log entry for ClickHouse
type LogEntry struct {
	Timestamp    time.Time         `ch:"timestamp"`
	TraceID      string            `ch:"trace_id"`
	Level        string            `ch:"level"`
	Service      string            `ch:"service"`
	Component    string            `ch:"component"`
	UserID       string            `ch:"user_id"`
	OrgID        string            `ch:"org_id"`
	WorkspaceID  string            `ch:"workspace_id"`
	Message      string            `ch:"message"`
	Fields       map[string]string `ch:"fields"`
	ErrorStack   string            `ch:"error_stack"`
	DurationMS   uint32            `ch:"duration_ms"`
	HTTPMethod   string            `ch:"http_method"`
	HTTPPath     string            `ch:"http_path"`
	HTTPStatus   uint16            `ch:"http_status"`
	SourceFile   string            `ch:"source_file"`
	SourceLine   uint32            `ch:"source_line"`
}

// AIOpsLogEntry represents AI operations log entry
type AIOpsLogEntry struct {
	Timestamp        time.Time         `ch:"timestamp"`
	TraceID          string            `ch:"trace_id"`
	Level            string            `ch:"level"`
	AgentType        string            `ch:"agent_type"`
	AgentID          string            `ch:"agent_id"`
	UserID           string            `ch:"user_id"`
	WorkspaceID      string            `ch:"workspace_id"`
	SessionID        string            `ch:"session_id"`
	Message          string            `ch:"message"`
	LLMModel         string            `ch:"llm_model"`
	PromptTokens     uint32            `ch:"prompt_tokens"`
	CompletionTokens uint32            `ch:"completion_tokens"`
	TotalTokens      uint32            `ch:"total_tokens"`
	LatencyMS        uint32            `ch:"latency_ms"`
	CostUSD          float32           `ch:"cost_usd"`
	Operation        string            `ch:"operation"`
	Status           string            `ch:"status"`
	Fields           map[string]string `ch:"fields"`
}

// PipelineLogEntry represents CI/CD pipeline log entry
type PipelineLogEntry struct {
	Timestamp   time.Time         `ch:"timestamp"`
	TraceID     string            `ch:"trace_id"`
	WorkspaceID string            `ch:"workspace_id"`
	ProjectID   string            `ch:"project_id"`
	PipelineID  string            `ch:"pipeline_id"`
	RunID       string            `ch:"run_id"`
	Stage       string            `ch:"stage"`
	Task        string            `ch:"task"`
	Level       string            `ch:"level"`
	Message     string            `ch:"message"`
	ExitCode    *int32            `ch:"exit_code"`
	DurationMS  *uint32           `ch:"duration_ms"`
	Provider    string            `ch:"provider"`
	Fields      map[string]string `ch:"fields"`
}

// SecurityLogEntry represents security event log entry
type SecurityLogEntry struct {
	Timestamp   time.Time         `ch:"timestamp"`
	EventType   string            `ch:"event_type"`
	Severity    string            `ch:"severity"`
	UserID      string            `ch:"user_id"`
	OrgID       string            `ch:"org_id"`
	WorkspaceID string            `ch:"workspace_id"`
	SourceIP    string            `ch:"source_ip"`
	UserAgent   string            `ch:"user_agent"`
	Resource    string            `ch:"resource"`
	Action      string            `ch:"action"`
	Result      string            `ch:"result"`
	Message     string            `ch:"message"`
	Metadata    map[string]string `ch:"metadata"`
	RiskScore   float32           `ch:"risk_score"`
}

// NewClickHouseLogger creates a new ClickHouse logger instance
func NewClickHouseLogger(config *ClickHouseConfig) (*ClickHouseLogger, error) {
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      time.Second * 30,
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}

	if config.TLS {
		options.TLS = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseLogger{
		conn:   conn,
		config: config,
	}, nil
}

// LogControlPlane logs control plane events
func (ch *ClickHouseLogger) LogControlPlane(ctx context.Context, entry *LogEntry) error {
	query := `INSERT INTO control_plane_logs (
		timestamp, trace_id, level, service, component, user_id, org_id, workspace_id,
		message, fields, error_stack, duration_ms, http_method, http_path, http_status,
		source_file, source_line
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := ch.conn.Exec(ctx, query,
		entry.Timestamp, entry.TraceID, entry.Level, entry.Service, entry.Component,
		entry.UserID, entry.OrgID, entry.WorkspaceID, entry.Message, entry.Fields,
		entry.ErrorStack, entry.DurationMS, entry.HTTPMethod, entry.HTTPPath,
		entry.HTTPStatus, entry.SourceFile, entry.SourceLine,
	)

	if err != nil {
		return fmt.Errorf("failed to insert control plane log: %w", err)
	}

	return nil
}

// LogAIOps logs AI operations events
func (ch *ClickHouseLogger) LogAIOps(ctx context.Context, entry *AIOpsLogEntry) error {
	query := `INSERT INTO aiops_logs (
		timestamp, trace_id, level, agent_type, agent_id, user_id, workspace_id, session_id,
		message, llm_model, prompt_tokens, completion_tokens, total_tokens, latency_ms,
		cost_usd, operation, status, fields
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := ch.conn.Exec(ctx, query,
		entry.Timestamp, entry.TraceID, entry.Level, entry.AgentType, entry.AgentID,
		entry.UserID, entry.WorkspaceID, entry.SessionID, entry.Message, entry.LLMModel,
		entry.PromptTokens, entry.CompletionTokens, entry.TotalTokens, entry.LatencyMS,
		entry.CostUSD, entry.Operation, entry.Status, entry.Fields,
	)

	if err != nil {
		return fmt.Errorf("failed to insert AIOps log: %w", err)
	}

	return nil
}

// LogPipeline logs CI/CD pipeline events
func (ch *ClickHouseLogger) LogPipeline(ctx context.Context, entry *PipelineLogEntry) error {
	query := `INSERT INTO pipeline_logs (
		timestamp, trace_id, workspace_id, project_id, pipeline_id, run_id, stage, task,
		level, message, exit_code, duration_ms, provider, fields
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := ch.conn.Exec(ctx, query,
		entry.Timestamp, entry.TraceID, entry.WorkspaceID, entry.ProjectID,
		entry.PipelineID, entry.RunID, entry.Stage, entry.Task, entry.Level,
		entry.Message, entry.ExitCode, entry.DurationMS, entry.Provider, entry.Fields,
	)

	if err != nil {
		return fmt.Errorf("failed to insert pipeline log: %w", err)
	}

	return nil
}

// LogSecurity logs security events
func (ch *ClickHouseLogger) LogSecurity(ctx context.Context, entry *SecurityLogEntry) error {
	query := `INSERT INTO security_events (
		timestamp, event_type, severity, user_id, org_id, workspace_id, source_ip,
		user_agent, resource, action, result, message, metadata, risk_score
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := ch.conn.Exec(ctx, query,
		entry.Timestamp, entry.EventType, entry.Severity, entry.UserID, entry.OrgID,
		entry.WorkspaceID, entry.SourceIP, entry.UserAgent, entry.Resource, entry.Action,
		entry.Result, entry.Message, entry.Metadata, entry.RiskScore,
	)

	if err != nil {
		return fmt.Errorf("failed to insert security log: %w", err)
	}

	return nil
}

// BatchLogControlPlane logs multiple control plane events in a batch
func (ch *ClickHouseLogger) BatchLogControlPlane(ctx context.Context, entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	batch, err := ch.conn.PrepareBatch(ctx, `INSERT INTO control_plane_logs (
		timestamp, trace_id, level, service, component, user_id, org_id, workspace_id,
		message, fields, error_stack, duration_ms, http_method, http_path, http_status,
		source_file, source_line
	)`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, entry := range entries {
		err := batch.Append(
			entry.Timestamp, entry.TraceID, entry.Level, entry.Service, entry.Component,
			entry.UserID, entry.OrgID, entry.WorkspaceID, entry.Message, entry.Fields,
			entry.ErrorStack, entry.DurationMS, entry.HTTPMethod, entry.HTTPPath,
			entry.HTTPStatus, entry.SourceFile, entry.SourceLine,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// QueryLogs executes a query on logs and returns results
func (ch *ClickHouseLogger) QueryLogs(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ch.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.ScanStruct(&row); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}

// Close closes the ClickHouse connection
func (ch *ClickHouseLogger) Close() error {
	if ch.conn != nil {
		return ch.conn.Close()
	}
	return nil
}

// Health checks the health of the ClickHouse connection
func (ch *ClickHouseLogger) Health(ctx context.Context) error {
	return ch.conn.Ping(ctx)
}