# hks-func CLI

The official CLI tool for deploying and managing serverless functions on the Hexabase AI platform.

## Features

- 🚀 **Easy Deployment** - Deploy functions with a single command
- 🔧 **Local Development** - Test functions locally before deployment
- 📦 **Multiple Runtimes** - Support for Node.js, Python, Go, Java, .NET, Ruby, PHP, and Rust
- 🔄 **Hot Reload** - Automatic reloading during local development
- 📊 **Real-time Logs** - Stream logs from deployed functions
- 🎯 **Traffic Management** - Control traffic routing between versions
- 🔐 **Secure** - Built-in authentication and secret management

## Installation

### Using Homebrew (macOS/Linux)

```bash
brew tap hexabase/tap
brew install hks-func
```

### Using Go

```bash
go install github.com/hexabase/hexabase-ai/cli/hks-func@latest
```

### Binary Releases

Download the latest binary from the [releases page](https://github.com/hexabase/hexabase-ai/releases).

## Quick Start

### 1. Configure Authentication

```bash
hks-func config auth
```

### 2. Initialize a New Function

```bash
# Interactive mode
hks-func init my-function

# With options
hks-func init my-function --runtime python --template http
```

### 3. Test Locally

```bash
cd my-function
hks-func test
```

### 4. Deploy

```bash
hks-func deploy
```

## Commands

### `init` - Initialize a new function

```bash
hks-func init [name] [flags]

Flags:
  -r, --runtime string    function runtime (node|python|go|java|dotnet|ruby|php|rust)
  -t, --template string   function template (http|event|scheduled|stream)
```

### `deploy` - Deploy function to platform

```bash
hks-func deploy [flags]

Flags:
  -n, --namespace string   target namespace
  -t, --tag string        image tag (default: latest)
      --dry-run           print deployment manifest without applying
      --force             force deployment even if no changes
      --no-traffic        deploy without routing traffic
```

### `list` - List deployed functions

```bash
hks-func list [flags]

Flags:
  -n, --namespace string   namespace to list from (default: default)
  -A, --all-namespaces    list functions from all namespaces
      --limit int         maximum number of functions to list
```

### `describe` - Show function details

```bash
hks-func describe [function-name] [flags]

Flags:
  -n, --namespace string   function namespace
```

### `logs` - View function logs

```bash
hks-func logs [function-name] [flags]

Flags:
  -n, --namespace string   function namespace
  -f, --follow            follow log output
      --tail int          number of recent lines to show
      --since string      show logs since duration (e.g. 10m, 1h)
  -c, --container string  specific container to get logs from
```

### `invoke` - Invoke a function

```bash
hks-func invoke [function-name] [flags]

Flags:
  -n, --namespace string   function namespace
  -d, --data string       JSON data to send
  -f, --file string       file containing data
  -m, --method string     HTTP method (default: POST)
  -p, --path string       request path (default: /)
  -H, --header strings    custom headers
      --async             invoke asynchronously
```

### `test` - Test function locally

```bash
hks-func test [flags]

Flags:
  -p, --port int          local port to run on (default: 8080)
      --env-file string   environment file to load
  -w, --watch            watch for changes and reload
```

### `build` - Build container image

```bash
hks-func build [flags]

Flags:
  -t, --tag string        image tag (default: latest)
      --push              push image to registry
      --platform string   target platform
```

### `delete` - Delete a function

```bash
hks-func delete [function-name] [flags]

Flags:
  -n, --namespace string   function namespace
  -f, --force             skip confirmation prompt
```

### `config` - Manage configuration

```bash
# Configure authentication
hks-func config auth

# Set configuration value
hks-func config set [key] [value]

# Get configuration value
hks-func config get [key]

# View all configuration
hks-func config view
```

## Function Configuration

Functions are configured using `function.yaml`:

```yaml
name: my-function
runtime: python
handler: main.handler
description: My serverless function
version: 1.0.0

environment:
  LOG_LEVEL: info
  API_KEY: ${SECRET_API_KEY}

build:
  builder: pack

deploy:
  namespace: production
  registry: registry.hexabase.ai/myorg
  autoscaling:
    minScale: 0
    maxScale: 100
    target: 100
    metric: concurrency
  resources:
    memory: 256Mi
    cpu: 100m
  timeout: 300
```

## Supported Runtimes

| Runtime | Versions | Builder | Handler Format |
|---------|----------|---------|----------------|
| Node.js | 14, 16, 18, 20 | pack | `index.handler` |
| Python | 3.8, 3.9, 3.10, 3.11 | pack | `main.handler` |
| Go | 1.19, 1.20, 1.21 | ko | `main` |
| Java | 11, 17, 21 | pack | `com.example.Handler` |
| .NET | 6.0, 7.0, 8.0 | pack | `Function::Handler` |
| Ruby | 3.0, 3.1, 3.2 | pack | `handler.rb` |
| PHP | 8.0, 8.1, 8.2 | pack | `index.php` |
| Rust | latest stable | docker | `handler` |

## Templates

### HTTP Function
Basic HTTP endpoint that responds to requests:
```bash
hks-func init my-api --template http
```

### Event Function
Processes CloudEvents from various sources:
```bash
hks-func init my-processor --template event
```

### Scheduled Function
Runs on a cron schedule:
```bash
hks-func init my-job --template scheduled
```

### Stream Function
Processes streaming data:
```bash
hks-func init my-stream --template stream
```

## Environment Variables

The CLI uses these environment variables:

- `HKS_API_TOKEN` - API authentication token
- `HKS_API_ENDPOINT` - API endpoint (default: https://api.hexabase.ai)
- `HKS_FUNC_NAMESPACE` - Default namespace
- `HKS_FUNC_REGISTRY` - Default container registry

## Examples

### Deploy a Python Function

```bash
# Initialize
hks-func init hello-world --runtime python --template http

# Develop
cd hello-world
# Edit main.py
hks-func test

# Deploy
hks-func deploy
```

### Blue-Green Deployment

```bash
# Deploy new version without traffic
hks-func deploy --tag v2 --no-traffic

# Test new version
hks-func invoke hello-world --tag v2

# Gradually shift traffic
hks-func traffic hello-world --tag v2 --percent 50

# Full cutover
hks-func traffic hello-world --tag v2 --percent 100
```

### Monitor Function

```bash
# View logs
hks-func logs hello-world -f

# Check status
hks-func describe hello-world

# View metrics (coming soon)
hks-func metrics hello-world
```

## Troubleshooting

### Authentication Issues

```bash
# Reconfigure authentication
hks-func config auth

# Check current config
hks-func config view
```

### Build Failures

```bash
# Use verbose output
hks-func build -v

# Try different builder
hks-func build --builder docker
```

### Deployment Issues

```bash
# Dry run to see manifest
hks-func deploy --dry-run

# Force deployment
hks-func deploy --force
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

Apache License 2.0