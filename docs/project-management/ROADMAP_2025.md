# Hexabase AI Roadmap 2025

**Last Updated**: 2025-06-09  
**Status**: Active Development

## 🎯 Project Vision

Transform Hexabase from a Kubernetes-as-a-Service platform into an AI-powered infrastructure management system that makes Kubernetes accessible to all developers through intelligent automation and natural language interfaces.

## 📊 Current Status

### ✅ Completed (as of Jan 2025)
- Core multi-tenant backend API (100% coverage)
- OAuth/OIDC authentication (Google, GitHub)
- Stripe billing integration
- VCluster lifecycle management
- Node & Application management APIs
- Basic AIOps integration
- Prometheus monitoring
- 650+ test cases

### 🔄 In Progress
- AIOps Python service (40% complete)
- Frontend UI (25% complete)
- Central logging with ClickHouse
- LLMOps with Langfuse

### ❌ Not Started
- Real LLM integration (Ollama)
- Email notifications
- CronJob management
- Serverless functions (HKS Functions)
- Multi-region support

## 🚀 Implementation Phases

### Phase 1: Core Platform & AI Agent Foundation (Jun-Jul 2025)

#### Week 1-2: CronJob Management Implementation
- [ ] Add CronJob type to Application model
- [ ] Create CronJob configuration API endpoints
- [ ] Implement Kubernetes CronJob resource creation
- [ ] Build UI for CronJob management
  - Schedule configuration (presets + cron expression)
  - Template selection from existing applications
  - Command/args override
  - Manual trigger capability
- [ ] Add job history and log viewing
- [ ] Integrate with backup settings feature

#### Week 3-4: HKS Functions (Serverless) Foundation
- [ ] Install Knative on host K3s cluster
- [ ] Develop hks-func CLI wrapper
  - Authentication integration
  - Project context management
  - Function deployment commands
- [ ] Create function management API
- [ ] Build UI dashboard for functions
  - Function listing and status
  - Invocation endpoints
  - Logs and metrics viewing
- [ ] Implement HTTP trigger routing
- [ ] Create function-to-CronJob integration

### Phase 2: AI Agent Enablement & Security (Jul-Aug 2025)

#### AI-Powered Dynamic Function Execution
- [ ] Implement secure Internal Operations API
  - `/internal/v1/operations/deploy-function`
  - `/internal/v1/operations/delete-function`
- [ ] Setup Kaniko for secure in-cluster builds
- [ ] Create restricted Service Accounts for functions
- [ ] Develop HKS Internal SDK (Python)
  - `hks_sdk.functions.execute(code="...")`
  - Automatic cleanup lifecycle
- [ ] Implement function sandboxing and resource limits
- [ ] Create AI agent code execution patterns

#### AIOps Integration with Functions
- [ ] Connect Ollama LLM service
- [ ] Enable AI agents to generate and deploy functions
- [ ] Implement Redis session management for context
- [ ] Setup ClickHouse for execution history
- [ ] Create monitoring & operations agents using functions

#### Security Enhancements
- [ ] JWT refresh token implementation
- [ ] Token revocation mechanism
- [ ] Session management & timeouts
- [ ] Rate limiting (per user/IP)
- [ ] PKCE flow in frontend
- [ ] Security headers (HSTS, CSP, etc.)

#### Frontend Development
- [ ] **Projects Management UI**
  - Project CRUD operations
  - Namespace visualization
  - Resource quota management
  - HNC hierarchy display
  
- [ ] **Billing Dashboard**
  - Subscription management
  - Payment methods
  - Invoice history
  - Usage analytics

- [ ] **Advanced Features**
  - RBAC visual management
  - Monitoring dashboards
  - Audit trail viewer
  - Group management

### Phase 3: Platform Intelligence (Aug-Oct 2025)

#### Observability Stack
- [ ] Multi-tenant Prometheus setup
- [ ] Grafana dashboard templates
- [ ] Alert rule configuration
- [ ] Log aggregation pipeline
- [ ] Distributed tracing

#### Advanced AIOps
- [ ] Predictive scaling algorithms
- [ ] Cost optimization recommendations
- [ ] Anomaly detection system
- [ ] Automated remediation workflows
- [ ] Change impact analysis

### Phase 4: Advanced Platform Features (Oct-Dec 2025)

#### Email & Notification Services
- [ ] Email notification system
  - Organization invitations
  - Billing notifications
  - CronJob execution alerts
  - Function invocation notifications
- [ ] Webhook integrations
- [ ] Slack/Discord notifications

#### NATS Message Queue Integration
- [ ] Async task processing
- [ ] Event-driven function triggers
- [ ] Retry mechanisms
- [ ] Dead letter queues

### Phase 5: Enterprise & Scale (2026 Q1-Q2)

#### Multi-Region Architecture
- [ ] Region-aware scheduling
- [ ] Cross-region replication
- [ ] Failover mechanisms
- [ ] Geo-distributed storage
- [ ] Region selection UI

#### Enterprise Features
- [ ] SAML/LDAP authentication
- [ ] Compliance certifications (SOC2, ISO)
- [ ] Data residency controls
- [ ] Enterprise SLAs
- [ ] White-label capabilities

## 📈 Success Metrics

### Q1 2025 Targets
- ✓ 100% API test coverage maintained
- ✓ <100ms API response time (p95)
- ✓ 0 critical security vulnerabilities
- ◯ 10+ active beta organizations

### Q2 2025 Targets
- ◯ 50+ active organizations
- ◯ 200+ workspaces deployed
- ◯ 99.9% platform uptime
- ◯ Complete UI feature parity

### Q3-Q4 2025 Targets
- ◯ 100+ active organizations
- ◯ 500+ workspaces deployed
- ◯ 1M+ API requests/day
- ◯ 50+ community contributions

## 🔧 Technical Debt Priorities

### Immediate Fixes (This Week)
1. Wire DI compilation issues
2. Helm RESTClientGetter implementation
3. Re-enable OAuth security tests
4. Repository name updates (kaas → ai)

### Short-term Improvements
1. OpenAPI/Swagger documentation
2. Error handling standardization
3. Performance benchmarks
4. Integration test suite
5. Logging consistency

### Long-term Architecture
1. Backup/restore procedures
2. Disaster recovery plan
3. Auto-scaling policies
4. Security scanning pipeline
5. Multi-cluster management

## 🚦 Risk Management

### High Priority Risks
1. **Kubernetes complexity** → Extensive documentation & abstractions
2. **Multi-tenancy security** → Regular audits & penetration testing
3. **Scalability limits** → Load testing & optimization
4. **LLM reliability** → Fallback mechanisms & caching

### Mitigation Strategies
- Weekly security reviews
- Monthly performance audits
- Continuous integration testing
- Community feedback loops
- Regular architecture reviews

## 🎯 Next Sprint Plan (Jun 9-23, 2025)

### Sprint Goals
1. Complete AIOps Ollama integration
2. Implement email notifications
3. Setup message queue processing
4. Begin security enhancements
5. Start Projects UI development

### Deliverables
- [ ] Working AI chat with real LLM
- [ ] Email service with 3+ templates
- [ ] Async task processing system
- [ ] JWT refresh token support
- [ ] Projects listing UI component

### Team Allocation
- **Backend**: AIOps, Email, Security
- **Frontend**: Projects UI, Auth flow
- **DevOps**: ClickHouse, Monitoring
- **QA**: Integration tests, Security audit

## 📞 Questions or Feedback?

Join our [GitHub Discussions](https://github.com/hexabase/hexabase-ai/discussions) or create an [Issue](https://github.com/hexabase/hexabase-ai/issues) for feature requests.

---

*This roadmap is a living document and will be updated based on community feedback and project progress.*