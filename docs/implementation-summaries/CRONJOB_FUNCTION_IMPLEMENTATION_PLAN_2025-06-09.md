# CronJob and Function Service Implementation Plan

**Date**: 2025-06-09  
**Priority**: HIGH  
**Sprint**: Jun 9-23, 2025

## Executive Summary

Implementing CronJob management and HKS Functions (serverless platform) as top priority features. These enable users to build AI agents, automated tasks, and event-driven applications with powerful capabilities including:

- Scheduled task execution
- Serverless function deployment
- AI-powered dynamic code execution
- Integration with backup settings
- Event-driven architectures

## 🎯 Implementation Roadmap

### Phase 1: CronJob Management (Days 1-4)

#### Backend Implementation

1. **Database Schema** ✅ Required
```sql
-- Extend applications table
ALTER TABLE applications 
ADD COLUMN type VARCHAR(20) DEFAULT 'stateless' CHECK (type IN ('stateless', 'cronjob', 'function')),
ADD COLUMN cron_schedule VARCHAR(100),
ADD COLUMN cron_command TEXT[],
ADD COLUMN cron_args TEXT[],
ADD COLUMN template_app_id UUID REFERENCES applications(id),
ADD COLUMN last_execution_at TIMESTAMP,
ADD COLUMN next_execution_at TIMESTAMP;

-- Job execution history
CREATE TABLE cronjob_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    job_name VARCHAR(255),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('running', 'succeeded', 'failed')),
    exit_code INT,
    logs TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cronjob_executions_app_id ON cronjob_executions(application_id);
CREATE INDEX idx_cronjob_executions_started_at ON cronjob_executions(started_at DESC);
```

2. **Domain Models**
```go
// internal/domain/application/models.go
type ApplicationType string

const (
    AppTypeStateless ApplicationType = "stateless"
    AppTypeCronJob   ApplicationType = "cronjob"
    AppTypeFunction  ApplicationType = "function"
)

type CronJobConfig struct {
    Schedule      string   `json:"schedule"`      // Cron expression
    Command       []string `json:"command"`       // Override command
    Args          []string `json:"args"`          // Override args
    TemplateAppID string   `json:"templateAppId"` // Source application
}

type CronJobExecution struct {
    ID            string    `json:"id"`
    ApplicationID string    `json:"applicationId"`
    JobName       string    `json:"jobName"`
    StartedAt     time.Time `json:"startedAt"`
    CompletedAt   *time.Time `json:"completedAt,omitempty"`
    Status        string    `json:"status"`
    ExitCode      *int      `json:"exitCode,omitempty"`
    Logs          string    `json:"logs"`
}
```

3. **API Endpoints**
```go
// internal/api/handlers/cronjobs.go
func (h *ApplicationHandler) CreateCronJob(c *gin.Context)
func (h *ApplicationHandler) ListExecutions(c *gin.Context)
func (h *ApplicationHandler) GetExecutionLogs(c *gin.Context)
func (h *ApplicationHandler) TriggerManually(c *gin.Context)
func (h *ApplicationHandler) UpdateSchedule(c *gin.Context)

// Routes
cronJobs := applications.Group("/cronjobs")
{
    cronJobs.POST("/", h.CreateCronJob)
    cronJobs.GET("/:id/executions", h.ListExecutions)
    cronJobs.POST("/:id/trigger", h.TriggerManually)
    cronJobs.GET("/:id/executions/:execId/logs", h.GetExecutionLogs)
    cronJobs.PATCH("/:id/schedule", h.UpdateSchedule)
}
```

4. **Kubernetes Integration**
```go
// internal/repository/kubernetes/cronjob.go
func (r *Repository) CreateCronJob(ctx context.Context, app *application.Application) error {
    // Get template application for container spec
    templateApp, err := r.appRepo.GetByID(ctx, app.CronJobConfig.TemplateAppID)
    if err != nil {
        return err
    }

    cronJob := &batchv1.CronJob{
        ObjectMeta: metav1.ObjectMeta{
            Name:      app.Name,
            Namespace: app.ProjectID,
            Labels: map[string]string{
                "app.kubernetes.io/managed-by": "hexabase",
                "hexabase.io/app-id":          app.ID,
                "hexabase.io/app-type":        "cronjob",
            },
        },
        Spec: batchv1.CronJobSpec{
            Schedule: app.CronJobConfig.Schedule,
            JobTemplate: batchv1.JobTemplateSpec{
                Spec: batchv1.JobSpec{
                    Template: corev1.PodTemplateSpec{
                        Spec: r.buildPodSpec(templateApp, app.CronJobConfig),
                    },
                },
            },
        },
    }
    
    return r.k8sClient.Create(ctx, cronJob)
}
```

#### Frontend Implementation

1. **UI Components**
```typescript
// ui/src/components/cronjob/cronjob-form.tsx
interface CronJobFormProps {
  projectId: string;
  applications: Application[]; // For template selection
  onSubmit: (data: CronJobData) => Promise<void>;
}

// ui/src/components/cronjob/schedule-builder.tsx
interface ScheduleBuilderProps {
  value: string;
  onChange: (schedule: string) => void;
}
// Presets: Hourly, Daily, Weekly, Monthly
// Advanced: Raw cron expression input

// ui/src/components/cronjob/execution-history.tsx
interface ExecutionHistoryProps {
  applicationId: string;
  executions: CronJobExecution[];
  onViewLogs: (executionId: string) => void;
}
```

### Phase 2: HKS Functions Platform (Days 5-8)

#### Infrastructure Setup

1. **Knative Installation**
```bash
# Install Knative Serving
kubectl apply -f https://github.com/knative/serving/releases/download/v1.13.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/v1.13.0/serving-core.yaml

# Install Kourier (ingress)
kubectl apply -f https://github.com/knative/net-kourier/releases/download/v1.13.0/kourier.yaml
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

# Configure DNS
kubectl apply -f deployments/k8s/knative/serving-default-domain.yaml
```

2. **Function Management API**
```go
// internal/domain/function/models.go
type Function struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    ProjectID   string            `json:"projectId"`
    Runtime     string            `json:"runtime"` // python, node, go
    Handler     string            `json:"handler"`
    Source      string            `json:"source"`  // Base64 encoded
    Environment map[string]string `json:"environment"`
    URL         string            `json:"url"`     // Knative service URL
    Status      string            `json:"status"`
    CreatedAt   time.Time         `json:"createdAt"`
}

// internal/api/handlers/functions.go
func (h *FunctionHandler) DeployFunction(c *gin.Context)
func (h *FunctionHandler) ListFunctions(c *gin.Context)
func (h *FunctionHandler) GetFunction(c *gin.Context)
func (h *FunctionHandler) DeleteFunction(c *gin.Context)
func (h *FunctionHandler) GetFunctionLogs(c *gin.Context)
```

3. **hks-func CLI Tool**
```go
// cli/cmd/hks-func/main.go
var rootCmd = &cobra.Command{
    Use:   "hks-func",
    Short: "HKS Functions CLI",
}

var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Deploy a function",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Build function image
        // 2. Push to registry
        // 3. Call HKS API to deploy
    },
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List functions in current project",
}
```

### Phase 3: AI Agent Integration (Days 9-10)

#### Dynamic Function Execution

1. **Internal Operations API**
```go
// internal/api/handlers/internal_operations.go
type DeployFunctionRequest struct {
    Code     string            `json:"code"`
    Runtime  string            `json:"runtime"`
    Handler  string            `json:"handler"`
    Env      map[string]string `json:"env"`
    Timeout  int               `json:"timeout"` // seconds
}

func (h *InternalHandler) DeployFunction(c *gin.Context) {
    // 1. Validate internal JWT
    // 2. Create temporary function with Kaniko
    // 3. Deploy to user's namespace
    // 4. Return function URL
}
```

2. **Python SDK**
```python
# sdk/python/hks_sdk/functions.py
class FunctionExecutor:
    def __init__(self, api_url: str, auth_token: str):
        self.api_url = api_url
        self.auth_token = auth_token
    
    def execute(self, code: str, input: dict = None, timeout: int = 30) -> dict:
        """Execute code as a serverless function"""
        # 1. Deploy function via Internal API
        deploy_resp = self._deploy_function(code, timeout)
        function_url = deploy_resp['url']
        function_id = deploy_resp['id']
        
        try:
            # 2. Invoke function
            result = self._invoke_function(function_url, input)
            return result
        finally:
            # 3. Cleanup
            self._delete_function(function_id)
```

#### Integration Examples

1. **Backup Automation**
```python
# AI agent generates backup function
backup_code = """
import boto3
import os

def handler(event):
    # Get database credentials from event
    db_config = event['db_config']
    s3_bucket = event['s3_bucket']
    
    # Perform backup
    backup_file = perform_backup(db_config)
    
    # Upload to S3
    s3 = boto3.client('s3')
    s3.upload_file(backup_file, s3_bucket, f"backups/{event['timestamp']}.sql")
    
    return {"status": "success", "file": backup_file}
"""

# Execute via CronJob
result = hks_sdk.functions.execute(
    code=backup_code,
    input={
        "db_config": {...},
        "s3_bucket": "my-backups",
        "timestamp": datetime.now().isoformat()
    }
)
```

2. **Monitoring Alert Handler**
```python
# AI generates alert handler based on user requirements
alert_handler = """
def handler(event):
    alert = event['alert']
    
    if alert['severity'] == 'critical':
        # Scale up replicas
        scale_deployment(alert['deployment'], replicas=5)
        
        # Notify on-call
        send_slack_message(alert['message'])
    
    return {"action": "handled"}
"""
```

## 🔧 Technical Considerations

### Security
- Functions run with restricted ServiceAccounts
- Network policies limit function access
- Resource quotas prevent resource exhaustion
- Code scanning before deployment

### Performance
- Function cold start optimization
- Shared runtime pools
- Autoscaling configuration
- Caching for frequently used functions

### Observability
- Structured logging to ClickHouse
- Prometheus metrics for functions
- Distributed tracing with OpenTelemetry
- Cost tracking per function

## 📊 Success Metrics

1. **Week 1 Goals**
   - [ ] CronJob CRUD operations working
   - [ ] UI for CronJob management complete
   - [ ] Manual trigger functionality
   - [ ] Execution history with logs

2. **Week 2 Goals**
   - [ ] Knative successfully installed
   - [ ] Functions deploying via API
   - [ ] hks-func CLI operational
   - [ ] AI agents executing dynamic functions

3. **Sprint Completion**
   - [ ] 10+ CronJobs created by beta users
   - [ ] 50+ function invocations
   - [ ] <2s function cold start time
   - [ ] 99% execution reliability

## 🚀 Next Steps After Sprint

1. **Advanced Features**
   - Event triggers (Kafka, NATS)
   - Function composition/chaining
   - Blue-green deployments
   - A/B testing for functions

2. **AI Enhancements**
   - Function recommendation engine
   - Automatic optimization suggestions
   - Cost prediction models
   - Performance profiling

3. **Enterprise Features**
   - Private function registry
   - Compliance scanning
   - Audit trails
   - SLA monitoring

This implementation will provide a powerful foundation for users to build sophisticated automation and AI-driven applications on the Hexabase platform.