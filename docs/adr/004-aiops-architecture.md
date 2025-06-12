# ADR-004: AI Operations (AIOps) Architecture

**Date**: 2025-06-06  
**Status**: Implemented  
**Authors**: AI Platform Team

## 1. Background

Hexabase AI needed to integrate AI/ML capabilities to provide:
- Intelligent automation for DevOps tasks
- Natural language interface for platform operations
- Predictive analytics for resource optimization
- Automated troubleshooting and remediation
- Context-aware assistance for developers

The challenge was creating a secure, scalable AI operations layer that could access platform resources while maintaining strict security boundaries.

## 2. Status

**Implemented** - Python-based AI operations service with Ollama for local LLM and support for external providers (OpenAI, Anthropic) is deployed.

## 3. Other Options Considered

### Option A: Direct LLM Integration
- Direct API calls to LLM providers
- Simple request/response model
- No intermediate processing layer

### Option B: Agent Framework with LangChain
- LangChain for orchestration
- Multiple specialized agents
- Tool calling capabilities

### Option C: Custom Agent Architecture with Sandbox
- Python-based agent system
- Secure sandbox execution
- Multi-provider LLM support
- Context management system

## 4. What Was Decided

We chose **Option C: Custom Agent Architecture with Sandbox** featuring:
- Python-based AIOps service running in isolated containers
- Secure sandbox model for agent execution
- Support for local (Ollama) and external LLMs
- Context management with user workspace awareness
- Tool registry for platform operations
- Comprehensive audit logging

## 5. Why Did You Choose It?

- **Security**: Sandbox isolation prevents unauthorized access
- **Flexibility**: Support multiple LLM providers
- **Control**: Custom architecture allows fine-tuned permissions
- **Performance**: Local LLM option for sensitive data
- **Scalability**: Containerized architecture scales horizontally

## 6. Why Didn't You Choose the Other Options?

### Why not Direct LLM Integration?
- No security boundaries
- Limited context management
- No tool calling capabilities
- Poor observability

### Why not LangChain?
- Too opinionated for our use case
- Security concerns with arbitrary code execution
- Difficult to customize for Kubernetes operations
- Heavy dependency footprint

## 7. What Has Not Been Decided

- Support for fine-tuned models
- Multi-modal capabilities (image/video analysis)
- Distributed agent coordination
- Real-time learning from user interactions

## 8. Considerations

### Architecture Overview
```
┌──────────────────┐
│   API Gateway    │
└────────┬─────────┘
         │
┌────────▼─────────┐
│  AIOps Service   │
│   (Python)       │
└────────┬─────────┘
         │
┌────────┴─────────┐
│  Agent Manager   │
├──────────────────┤
│ Security Sandbox │
├──────────────────┤
│ Context Manager  │
└──────┬─────┬────┘
       │     │
┌──────▼─┐ ┌─▼──────────┐
│ Ollama │ │External LLM│
└────────┘ └────────────┘
```

### Security Model
```python
class SecuritySandbox:
    def __init__(self, user_context):
        self.user_id = user_context.user_id
        self.org_id = user_context.org_id
        self.permissions = self._load_permissions()
    
    def execute_tool(self, tool_name, params):
        if not self._check_permission(tool_name):
            raise PermissionDenied()
        
        # Execute in isolated environment
        with sandboxed_execution():
            return self.tool_registry[tool_name](**params)
```

### Tool Registry
- **Kubernetes Operations**: List/describe resources, get logs
- **Metric Analysis**: Query Prometheus, analyze trends
- **Code Generation**: Generate YAML, scripts, configurations
- **Troubleshooting**: Analyze errors, suggest fixes

### Context Management
- User workspace and project context
- Historical conversation memory
- Resource access patterns
- Performance baselines

### LLM Provider Configuration
```python
providers = {
    "ollama": {
        "endpoint": "http://ollama:11434",
        "model": "mistral",
        "timeout": 30
    },
    "openai": {
        "endpoint": "https://api.openai.com/v1",
        "model": "gpt-4-turbo",
        "api_key": "${OPENAI_API_KEY}"
    }
}
```

### Performance Considerations
- Response streaming for better UX
- Caching for repeated queries
- Async execution for long-running tasks
- Connection pooling for LLM providers

### Compliance and Audit
- All AI interactions logged
- PII detection and masking
- Model decision explanations
- Usage tracking and quotas

### Future Enhancements
- RAG (Retrieval Augmented Generation) for documentation
- Multi-agent collaboration
- Automated workflow generation
- Continuous learning from outcomes