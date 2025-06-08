# Hexabase AI - Per-User AI Agent Architecture

**Created**: 2025-06-08  
**Feature**: User-Scoped AI Agents with Cluster Credentials

## 🎯 Overview

The Hexabase AI platform provides each logged-in user with a personal AI agent that operates with their specific Kubernetes cluster credentials. This allows users to interact with their workspaces through natural language instead of complex UI operations.

## 🔑 Key Concepts

### User-Scoped Agent
- **One agent per user session**: When a user logs in, an AI agent is initialized with their credentials
- **Workspace context**: Agent knows which workspaces, projects, and namespaces the user can access
- **RBAC enforcement**: Agent operates within the user's permission boundaries
- **Session persistence**: Agent maintains conversation context throughout the session

### Credential Flow
```
User Login → OAuth/JWT → Extract Workspace Access → Generate Kubeconfig → Initialize Agent
```

## 🏗️ Architecture

### 1. Agent Initialization

When a user accesses the chat interface (`/api/v1/ai/chat`):

```go
// Go API creates agent context
agentContext := AgentContext{
    UserID:      user.ID,
    WorkspaceID: workspace.ID,
    Kubeconfig:  workspace.GetKubeconfigForUser(user),
    Permissions: rbac.GetUserPermissions(user, workspace),
    Namespaces:  projects.GetUserNamespaces(user, workspace),
}

// Pass to Python AI service
internalToken := auth.GenerateInternalAIOpsToken(
    userID,
    orgIDs,
    workspaceID,
    agentContext,
)
```

### 2. Python AI Agent

```python
class UserClusterAgent:
    def __init__(self, user_context: UserContext):
        self.user_id = user_context.user_id
        self.workspace_id = user_context.workspace_id
        
        # Initialize K8s client with user's credentials
        self.k8s_client = self._init_k8s_client(user_context.kubeconfig)
        
        # Available tools based on permissions
        self.tools = self._init_tools_for_user(user_context.permissions)
```

### 3. Available Tools

**Read-Only Tools** (Always Available):
- `ListPodsInNamespace` - View pods and their status
- `GetDeploymentStatus` - Check deployment health
- `ViewLogs` - Read container logs
- `GetResourceUsage` - Monitor CPU/memory usage
- `ListServices` - View service endpoints
- `GetEvents` - Check Kubernetes events

**Write Tools** (Require Confirmation):
- `ScaleDeployment` - Change replica count
- `RestartPod` - Restart problematic pods
- `CreateCronJob` - Schedule tasks
- `UpdateConfigMap` - Modify configurations
- `CreateBackupJob` - Initiate backups
- `ApplyYAML` - Deploy resources

**AI-Specific Tools**:
- `AnalyzePerformance` - AI-driven performance analysis
- `SuggestOptimizations` - Resource optimization recommendations
- `GenerateBackupStrategy` - Intelligent backup planning
- `TroubleshootIssues` - Automated problem diagnosis

## 💬 Interaction Examples

### Example 1: Checking Application Status
```
User: "How is my web app doing?"

Agent: "I'll check the status of your web application. Let me examine the deployments 
and pods in your project namespace...

✅ Deployment 'web-app' is healthy:
- 3/3 replicas running
- All pods ready
- CPU usage: 45% average
- Memory usage: 62% average
- No recent errors in logs

The application is performing well with no issues detected."
```

### Example 2: Scaling with Confirmation
```
User: "My API is slow, can you scale it up?"

Agent: "I've analyzed your API deployment. Current status:
- Running 2 replicas
- CPU usage: 85% (high)
- Memory usage: 70%
- Request latency: 450ms average

I recommend scaling to 4 replicas to handle the load better.

⚠️ This action will scale 'api-service' from 2 to 4 replicas. 
Estimated cost increase: ~$15/month

Do you want me to proceed? [Yes/No]"

User: "Yes"

Agent: "Scaling deployment... Done! 
- New replicas starting up
- 4/4 replicas ready in 45 seconds
- CPU usage dropping to 45%
- Latency improved to 120ms"
```

### Example 3: Intelligent Backup Creation
```
User: "Set up backups for my database"

Agent: "I'll analyze your database setup and create a backup strategy...

Found PostgreSQL deployment 'user-db' with:
- Data volume: 50GB
- Daily change rate: ~2GB
- Current storage class: fast-ssd

Recommended backup strategy:
📅 Schedule: Daily at 2 AM (low traffic period)
💾 Retention: 30 days (compliance-friendly)
🗄️ Type: Incremental with weekly full backups
☁️ Storage: S3 bucket in same region

I'll create a CronJob that:
1. Takes consistent snapshots using pg_dump
2. Compresses with gzip (est. 70% reduction)
3. Encrypts with AES-256
4. Uploads to S3 with lifecycle policies

Shall I implement this backup strategy? [Yes/No]"
```

### Example 4: Troubleshooting
```
User: "My app keeps crashing"

Agent: "I'll investigate the crashes. Analyzing recent events and logs...

🔍 Found the issue:
- Pod 'app-6b7d4-x2n4' has restarted 12 times
- Error: OOMKilled (Out of Memory)
- Current limit: 512Mi
- Peak usage before crash: 580Mi

Root cause: Memory limit too low for current workload

Recommended fixes:
1. Immediate: Increase memory limit to 1Gi
2. Long-term: Implement memory profiling
3. Optional: Enable HPA for auto-scaling

Would you like me to apply the memory limit increase? [Yes/No]"
```

## 🔒 Security Features

### 1. Permission Boundaries
```python
async def execute_action(self, action: str, params: dict):
    # Check RBAC permissions
    if not self.has_permission(action):
        return "Sorry, you don't have permission to perform this action."
    
    # High-risk actions need confirmation
    if self.is_high_risk_action(action):
        confirmed = await self.request_confirmation(
            f"This will {self.describe_action(action)}. Proceed?"
        )
        if not confirmed:
            return "Action cancelled."
    
    # Execute with user's credentials
    return await self.execute_with_k8s_client(action, params)
```

### 2. Audit Logging
All agent actions are logged to ClickHouse:
- User ID and workspace
- Action performed
- Confirmation status
- Result/error
- Timestamp
- Resource affected

### 3. Session Isolation
- Each user's agent is isolated
- No cross-user data access
- Credentials expire with session
- Memory cleared on logout

## 🚀 Implementation Phases

### Phase 1: Basic Chat Integration (Week 1)
- Connect Ollama for LLM responses
- Implement session management
- Basic read-only Kubernetes tools

### Phase 2: User Context Integration (Week 2)
- Pass kubeconfig to agent
- Initialize K8s client per user
- Implement RBAC checking

### Phase 3: Interactive Tools (Week 3)
- Build confirmation flow
- Implement write operations
- Add safety checks

### Phase 4: Intelligent Features (Week 4)
- Performance analysis
- Backup strategies
- Troubleshooting automation
- Cost optimization

## 📈 Benefits

1. **Simplified Operations**: Users can manage Kubernetes with natural language
2. **Reduced Errors**: AI validates actions before execution
3. **Learning Tool**: Agent explains what it's doing, teaching Kubernetes concepts
4. **Accessibility**: Makes Kubernetes accessible to non-experts
5. **Audit Trail**: All actions tracked for compliance

## 🎯 Success Metrics

- 80% of common tasks completed via chat
- 90% reduction in configuration errors
- 50% faster issue resolution
- 95% user satisfaction score

This architecture ensures that each user has a powerful, secure, and personalized AI assistant for managing their Kubernetes workloads.