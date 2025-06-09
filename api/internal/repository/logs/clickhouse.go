package logs

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/hexabase/hexabase-ai/api/internal/domain/logs"
)

type ClickHouseRepository struct {
	conn clickhouse.Conn
}

func NewClickHouseRepository(conn clickhouse.Conn) logs.Repository {
	return &ClickHouseRepository{conn: conn}
}

func (r *ClickHouseRepository) QueryLogs(ctx context.Context, query logs.LogQuery) ([]logs.LogEntry, error) {
	var args []interface{}
	
	// Start with a base query
	sql := "SELECT timestamp, level, message, trace_id, user_id, source, details FROM logs WHERE workspace_id = ?"
	args = append(args, query.WorkspaceID)

	// Add conditions dynamically
	if query.SearchTerm != "" {
		sql += " AND message ILIKE ?"
		args = append(args, "%"+query.SearchTerm+"%")
	}
	if query.Level != "" {
		sql += " AND level = ?"
		args = append(args, query.Level)
	}
	if !query.StartTime.IsZero() {
		sql += " AND timestamp >= ?"
		args = append(args, query.StartTime)
	}
	if !query.EndTime.IsZero() {
		sql += " AND timestamp <= ?"
		args = append(args, query.EndTime)
	}
	
	sql += " ORDER BY timestamp DESC"

	if query.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, query.Limit)
	}

	rows, err := r.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute log query: %w", err)
	}
	defer rows.Close()

	var results []logs.LogEntry
	for rows.Next() {
		var entry logs.LogEntry
		// Details are complex to scan directly, this needs a proper implementation
		// For now, we will skip scanning 'details'
		var details string // placeholder
		if err := rows.Scan(&entry.Timestamp, &entry.Level, &entry.Message, &entry.TraceID, &entry.UserID, &entry.Source, &details); err != nil {
			return nil, fmt.Errorf("failed to scan log row: %w", err)
		}
		results = append(results, entry)
	}

	return results, nil
} 