# Immediate Next Steps - Sprint Plan

**Sprint**: Jun 9-23, 2025  
**Focus**: CronJob & Function Service Implementation

## 🎯 Priority 1: CronJob Management (3-4 days)

### 1. Database Schema Updates
```sql
-- Add to application model
ALTER TABLE applications ADD COLUMN type VARCHAR(20) DEFAULT 'stateless';
ALTER TABLE applications ADD COLUMN cron_schedule VARCHAR(100);
ALTER TABLE applications ADD COLUMN cron_command TEXT[];
ALTER TABLE applications ADD COLUMN cron_args TEXT[];
ALTER TABLE applications ADD COLUMN template_app_id UUID REFERENCES applications(id);

-- CronJob execution history
CREATE TABLE cronjob_executions (
    id UUID PRIMARY KEY,
    application_id UUID REFERENCES applications(id),
    job_name VARCHAR(255),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    status VARCHAR(20),
    logs TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. API Endpoints
```go
// New endpoints in /api/internal/api/handlers/applications.go
- POST   /applications/cronjob         // Create CronJob
- GET    /applications/:id/executions  // List job executions
- POST   /applications/:id/trigger     // Manual trigger
- GET    /applications/:id/executions/:execId/logs  // Get logs
```

### 3. Kubernetes CronJob Creation
```go
// New file: /api/internal/repository/kubernetes/cronjob.go
func (r *Repository) CreateCronJob(ctx context.Context, app *Application) error {
    cronJob := &batchv1.CronJob{
        ObjectMeta: metav1.ObjectMeta{
            Name:      app.Name,
            Namespace: app.ProjectID,
        },
        Spec: batchv1.CronJobSpec{
            Schedule: app.CronSchedule,
            JobTemplate: // ... pod spec from template app
        },
    }
    return r.k8sClient.Create(ctx, cronJob)
}
```

### 4. UI Components
```typescript
// New components in /ui/src/components/
- cronjob-form.tsx         // CronJob creation form
- schedule-picker.tsx      // User-friendly schedule UI
- execution-history.tsx    // Job execution history
- log-viewer.tsx          // Real-time log streaming
```

## 🎯 Priority 2: HKS Functions Foundation (4-5 days)

### 1. Knative Installation
```yaml
# Location: /deployments/k8s/knative/
- serving-core.yaml        # Knative Serving components
- serving-hpa.yaml         # Autoscaling configuration
- kourier.yaml            # Ingress controller
- serving-default-domain.yaml  # DNS configuration
```

### 2. hks-func CLI Development
```go
// New cmd: /cli/cmd/hks-func/
- main.go                  // CLI entry point
- auth.go                  // HKS authentication
- deploy.go               // Function deployment
- list.go                 // List functions
- logs.go                 // Stream function logs
```

### 3. Function Management API
```go
// New handlers: /api/internal/api/handlers/functions.go
- POST   /projects/:id/functions       // Deploy function
- GET    /projects/:id/functions       // List functions
- GET    /functions/:id                // Get function details
- DELETE /functions/:id                // Delete function
- GET    /functions/:id/logs           // Stream logs
```

### 4. Knative Service Creation
```go
// New file: /api/internal/repository/kubernetes/knative.go
func (r *Repository) DeployFunction(ctx context.Context, fn *Function) error {
    ksvc := &servingv1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fn.Name,
            Namespace: fn.ProjectID,
        },
        Spec: servingv1.ServiceSpec{
            Template: servingv1.RevisionTemplateSpec{
                Spec: servingv1.RevisionSpec{
                    Containers: []corev1.Container{{
                        Image: fn.Image,
                        Env:   convertEnvVars(fn.Environment),
                    }},
                },
            },
        },
    }
    return r.k8sClient.Create(ctx, ksvc)
}
```

## 🎯 Priority 3: AI Agent Integration (2-3 days)

### 1. Internal Operations API
```go
// New file: /api/internal/api/handlers/internal_operations.go
- POST /internal/v1/operations/deploy-function
- DELETE /internal/v1/operations/delete-function
- GET /internal/v1/operations/function-status/:id
```

### 2. Secure Function Builder
```go
// New package: /api/internal/builder/
- kaniko.go               // Kaniko builder integration
- sandbox.go              // Security sandbox configuration
- templates.go            // Function templates
```

### 3. HKS Internal SDK (Python)
```python
# New package: /sdk/python/hks_sdk/
- functions.py            # Function execution API
- auth.py                # Internal JWT handling
- client.py              # HTTP client wrapper

# Example usage:
import hks_sdk

result = hks_sdk.functions.execute(
    code="""
    def handler(event):
        return {"result": event["input"] * 2}
    """,
    input={"input": 21},
    timeout=30
)
```

### 4. Function Templates
```yaml
# Location: /api/internal/builder/templates/
- python-base.yaml        # Python function template
- node-base.yaml         # Node.js function template
- go-base.yaml          # Go function template
```

## 🎯 Priority 4: Integration & Testing (2-3 days)

### 1. CronJob + Function Integration
```go
// Enable CronJobs to trigger functions
type CronJobFunctionTrigger struct {
    FunctionURL string
    AuthToken   string
    Payload     map[string]interface{}
}
```

### 2. Backup Settings Integration
```go
// Add to Application model
type BackupSettings struct {
    Enabled      bool
    Schedule     string  // Cron expression
    Destination  string  // S3, GCS, etc.
    Retention    int     // Days
    FunctionID   string  // Custom backup function
}
```

### 3. End-to-End Tests
```go
// New test files:
- cronjob_integration_test.go
- function_deployment_test.go
- ai_agent_execution_test.go
```

### 4. Documentation
```markdown
# New docs:
- /docs/features/cronjobs.md
- /docs/features/functions.md
- /docs/tutorials/ai-agent-functions.md
- /docs/api-reference/cronjobs-api.md
- /docs/api-reference/functions-api.md
```

## 📋 Daily Checklist

### Day 1-2: CronJob Backend
- [ ] Update database schema for CronJobs
- [ ] Implement CronJob API endpoints
- [ ] Create Kubernetes CronJob resources
- [ ] Add execution history tracking

### Day 3-4: CronJob UI
- [ ] Build CronJob creation form
- [ ] Implement schedule picker component
- [ ] Create execution history view
- [ ] Add log viewer component

### Day 5-6: Knative Setup
- [ ] Install Knative on host cluster
- [ ] Configure Kourier ingress
- [ ] Setup DNS for functions
- [ ] Test basic function deployment

### Day 7-8: Function Management
- [ ] Implement function API endpoints
- [ ] Create hks-func CLI tool
- [ ] Build function dashboard UI
- [ ] Test function deployment flow

### Day 9-10: AI Agent Integration
- [ ] Implement Internal Operations API
- [ ] Setup Kaniko builder
- [ ] Create Python SDK
- [ ] Test AI-driven function execution

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

1. **Knative Installation**: Requires cluster admin access and resources
2. **DNS Configuration**: Need wildcard DNS for function URLs
3. **Container Registry**: Need registry for function images
4. **Security Policies**: May need to adjust PSPs/SCCs for builders

## 📊 Success Criteria

- [ ] Users can create and manage CronJobs via UI
- [ ] CronJobs execute reliably on schedule
- [ ] Functions deploy successfully via CLI and API
- [ ] AI agents can dynamically create and execute functions
- [ ] Integration with backup settings works end-to-end
- [ ] All new code has >90% test coverage

## 🔄 Daily Standup Topics

1. CronJob implementation progress
2. Knative setup challenges
3. Function deployment testing
4. AI agent integration status
5. Performance and security considerations

---

**Note**: This is a living document. Update daily with progress and adjust timelines as needed.