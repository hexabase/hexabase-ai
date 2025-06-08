# Immediate Next Steps - Sprint Plan

**Sprint**: Jan 8-22, 2025  
**Focus**: AIOps Completion & Core Services

## 🎯 Priority 1: Complete AIOps Integration (3-4 days)

### 1. Connect Real Ollama Service
```bash
# Location: /ai-ops/src/aiops/llm_clients/ollama.py
```
- [ ] Replace mock responses with actual Ollama API calls
- [ ] Implement proper error handling for LLM failures
- [ ] Add request timeout and retry logic
- [ ] Create Ollama deployment YAML for k8s

### 2. Redis Chat Session Management
```python
# New file: /ai-ops/src/aiops/core/session.py
class ChatSessionManager:
    def create_session(user_id: str, org_id: str) -> str
    def get_session(session_id: str) -> ChatSession
    def update_context(session_id: str, context: dict)
    def expire_session(session_id: str)
```

### 3. ClickHouse Chat History
```sql
-- Schema: /deployments/monitoring/clickhouse/schema.sql
CREATE TABLE chat_history (
    session_id UUID,
    timestamp DateTime,
    user_id String,
    org_id String,
    message_type Enum('user', 'assistant', 'system'),
    content String,
    metadata JSON
) ENGINE = MergeTree()
ORDER BY (org_id, user_id, timestamp);
```

### 4. Agent Implementation
```python
# Location: /ai-ops/src/aiops/agents/
- monitoring_agent.py    # Prometheus queries, alerts
- operations_agent.py    # Workspace operations
- troubleshooting_agent.py  # Log analysis, debugging
```

## 🎯 Priority 2: Email Notifications (2-3 days)

### 1. Email Service Setup
```go
// New package: /api/internal/email/
- service.go         // Core email service
- templates.go       // Email templates
- sender.go          // SMTP/SendGrid interface
```

### 2. Required Templates
- Organization invitation
- Workspace provisioning (started/completed/failed)
- Billing notifications (subscription/payment/invoice)
- Password reset (if local auth enabled)
- Alert notifications

### 3. Integration Points
```go
// Update existing services:
- OrganizationService.InviteMember() → Send invitation email
- WorkspaceService.Create() → Send provisioning notifications
- BillingService webhook handlers → Send payment emails
```

## 🎯 Priority 3: Message Queue Integration (1-2 days)

### 1. NATS Publisher
```go
// Location: /api/internal/messaging/publisher.go
type Publisher interface {
    PublishWorkspaceTask(task WorkspaceTask) error
    PublishEmailTask(email EmailTask) error
}
```

### 2. Worker Implementation
```go
// New cmd: /api/cmd/worker/main.go
- Subscribe to task queues
- Process workspace provisioning
- Handle email sending
- Implement retry logic
```

### 3. Task Status Tracking
```sql
-- Add to existing schema
CREATE TABLE async_tasks (
    id UUID PRIMARY KEY,
    type VARCHAR(50),
    status VARCHAR(20),
    payload JSONB,
    error TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## 🎯 Priority 4: Security Enhancements (3-4 days)

### 1. JWT Refresh Tokens
```go
// Update: /api/internal/auth/jwt.go
- Add refresh token generation
- Implement token rotation
- Add revocation checks
```

### 2. Rate Limiting
```go
// New middleware: /api/internal/api/middleware/ratelimit.go
- Per-user limits
- Per-IP limits  
- Endpoint-specific limits
```

### 3. Frontend PKCE Flow
```typescript
// Update: /ui/src/lib/auth-context.tsx
- Implement PKCE challenge/verifier
- Add secure token storage
- Handle refresh flow
```

## 📋 Daily Checklist

### Day 1-2: AIOps Foundation
- [ ] Deploy Ollama to k8s cluster
- [ ] Update OllamaClient implementation
- [ ] Create Redis session manager
- [ ] Write session management tests

### Day 3-4: AIOps Agents
- [ ] Implement monitoring agent
- [ ] Create operations agent
- [ ] Setup ClickHouse schema
- [ ] Test end-to-end chat flow

### Day 5-6: Email Service
- [ ] Setup email service structure
- [ ] Create email templates
- [ ] Integrate with organization service
- [ ] Test email delivery

### Day 7-8: Message Queue
- [ ] Implement NATS publisher
- [ ] Create worker process
- [ ] Add retry mechanisms
- [ ] Test async processing

### Day 9-10: Security
- [ ] Implement refresh tokens
- [ ] Add rate limiting
- [ ] Update frontend auth
- [ ] Security testing

## 🧪 Testing Requirements

### AIOps Tests
```bash
cd /ai-ops
pytest tests/test_ollama_integration.py
pytest tests/test_session_manager.py
pytest tests/test_agents.py
```

### Email Tests
```bash
cd /api
go test ./internal/email/...
go test ./internal/service/... -run TestEmailIntegration
```

### Security Tests
```bash
# Re-enable and update
go test ./internal/auth/oauth_security_test.go
go test ./internal/api/middleware/ratelimit_test.go
```

## 🚨 Potential Blockers

1. **Ollama Deployment**: Need GPU nodes or CPU-optimized models
2. **Email Provider**: Need SMTP credentials or SendGrid API key
3. **ClickHouse Setup**: Requires persistent storage configuration
4. **Redis Cluster**: May need Redis Sentinel for HA

## 📊 Success Criteria

- [ ] Users can have natural language conversations with AI
- [ ] All critical operations send email notifications
- [ ] Async tasks process reliably with retries
- [ ] Security improvements pass penetration testing
- [ ] All new code has >90% test coverage

## 🔄 Daily Standup Topics

1. AIOps integration progress
2. Blockers with external services
3. Test coverage status
4. Security review findings
5. Performance metrics

---

**Note**: This is a living document. Update daily with progress and adjust timelines as needed.