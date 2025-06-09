# Quick Start Guide

Get up and running with Hexabase AI in minutes.

## Prerequisites

Before you begin, ensure you have:
- A Hexabase AI account
- kubectl installed on your machine
- Basic knowledge of Kubernetes concepts

## 1. Sign Up and Login

### Create an Account

1. Visit [https://app.hexabase.ai](https://app.hexabase.ai)
2. Click "Sign Up"
3. Choose your authentication method:
   - Google OAuth
   - GitHub OAuth
   - Email/Password

### First Login

After signing up, you'll be guided through the initial setup:

1. **Create Organization**: Your billing and team management unit
2. **Choose Plan**: Select from Starter, Pro, or Enterprise
3. **Create First Workspace**: Your isolated Kubernetes environment

## 2. Create Your First Workspace

Workspaces are isolated Kubernetes environments (vClusters) for your applications.

### Using the UI

1. Navigate to the Workspaces page
2. Click "Create Workspace"
3. Fill in the details:
   ```
   Name: my-first-workspace
   Description: Getting started with Hexabase AI
   Plan: Starter (1 CPU, 2GB RAM)
   ```
4. Click "Create"

The workspace will be provisioned in about 30 seconds.

### Using the CLI

```bash
# Install Hexabase CLI
curl -sSL https://get.hexabase.ai | sh

# Login
hks login

# Create workspace
hks workspace create my-first-workspace --plan starter
```

## 3. Connect to Your Workspace

Once your workspace is ready, you can connect using kubectl.

### Download Kubeconfig

1. Go to your workspace dashboard
2. Click "Download Kubeconfig"
3. Save the file to `~/.kube/config-hexabase`

### Configure kubectl

```bash
# Set the kubeconfig
export KUBECONFIG=~/.kube/config-hexabase

# Verify connection
kubectl get nodes
```

## 4. Deploy Your First Application

Let's deploy a simple web application.

### Create a Project (Namespace)

Projects organize resources within a workspace.

```bash
# Using CLI
hks project create my-app --workspace my-first-workspace

# Or using kubectl
kubectl create namespace my-app
```

### Deploy an Application

Create a file named `app.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
  namespace: my-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
      - name: hello
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world
  namespace: my-app
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-world
  namespace: my-app
spec:
  rules:
  - host: hello.my-first-workspace.hexabase.app
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hello-world
            port:
              number: 80
```

Deploy the application:

```bash
kubectl apply -f app.yaml
```

### Access Your Application

Your application will be available at:
```
https://hello.my-first-workspace.hexabase.app
```

## 5. Manage Your Application

### View Application Status

```bash
# Check deployment status
kubectl get deployments -n my-app

# View pods
kubectl get pods -n my-app

# Check service
kubectl get svc -n my-app
```

### Scale Your Application

```bash
# Scale to 5 replicas
kubectl scale deployment hello-world -n my-app --replicas=5

# Or use the UI
# Navigate to Applications > hello-world > Scale
```

### View Logs

```bash
# View logs from all pods
kubectl logs -n my-app -l app=hello-world

# Stream logs
kubectl logs -n my-app -l app=hello-world -f
```

## 6. Set Up Monitoring

Hexabase AI includes built-in monitoring.

### Access Grafana Dashboard

1. Go to your workspace dashboard
2. Click "Monitoring"
3. View metrics for:
   - CPU usage
   - Memory usage
   - Network traffic
   - Request rates

### Set Up Alerts

1. Navigate to Monitoring > Alerts
2. Click "Create Alert"
3. Configure alert conditions:
   ```
   Metric: CPU Usage
   Condition: > 80%
   Duration: 5 minutes
   Action: Send email
   ```

## 7. Deploy a CronJob

Schedule recurring tasks using CronJobs.

### Create a Backup CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
  namespace: my-app
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: your-backup-image:latest
            command:
            - /bin/sh
            - -c
            - echo "Running backup..." && date
          restartPolicy: OnFailure
```

Deploy:

```bash
kubectl apply -f backup-cronjob.yaml
```

## 8. Use AI Assistant

Hexabase AI includes an AI assistant for troubleshooting.

### Ask for Help

```bash
# Using CLI
hks ai "Why is my deployment failing?"

# The AI will analyze your resources and provide suggestions
```

### Example AI Commands

- "Help me optimize my resource limits"
- "Why is my pod crashing?"
- "Generate a Dockerfile for a Node.js app"
- "Create a CI/CD pipeline for my application"

## 9. Manage Team Access

### Invite Team Members

1. Go to Organization Settings
2. Click "Team Members"
3. Click "Invite Member"
4. Enter email and select role:
   - **Admin**: Full access
   - **Developer**: Deploy and manage applications
   - **Viewer**: Read-only access

### Set Workspace Permissions

1. Navigate to Workspace > Settings > Access
2. Add team members
3. Assign roles specific to this workspace

## 10. Next Steps

### Explore Advanced Features

- **Serverless Functions**: Deploy event-driven functions
- **Backup & Restore**: Set up automated backups
- **CI/CD Integration**: Connect your Git repository
- **Custom Domains**: Use your own domain names
- **Multi-region Deployment**: Deploy across regions

### Learn More

- [Core Concepts](./concepts.md)
- [API Documentation](../../api-reference/README.md)
- [Video Tutorials](https://hexabase.ai/tutorials)
- [Community Forum](https://community.hexabase.ai)

### Get Help

- **Documentation**: [docs.hexabase.ai](https://docs.hexabase.ai)
- **Support**: support@hexabase.ai
- **Discord**: [Join our Discord](https://discord.gg/hexabase)
- **Status Page**: [status.hexabase.ai](https://status.hexabase.ai)

## Troubleshooting

### Common Issues

**Cannot connect to workspace**
```bash
# Verify kubeconfig
kubectl config current-context

# Test connection
kubectl cluster-info
```

**Application not accessible**
```bash
# Check ingress
kubectl get ingress -n my-app

# Verify DNS
nslookup hello.my-first-workspace.hexabase.app
```

**Out of resources**
```bash
# Check resource usage
kubectl top nodes
kubectl top pods -n my-app

# View quotas
kubectl get resourcequota -n my-app
```

---

🎉 **Congratulations!** You've successfully deployed your first application on Hexabase AI. Continue exploring our platform's features to build and scale your applications with ease.