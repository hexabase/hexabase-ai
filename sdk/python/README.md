# Hexabase AI Python SDK

Python SDK for deploying and executing serverless functions on the Hexabase AI platform.

## Features

- 🚀 **Easy Function Deployment**: Deploy Python functions with a single method call
- 🔄 **Auto-Cleanup**: Automatic cleanup of functions based on TTL, execution count, or idle time
- 🔐 **Secure Authentication**: Built-in authentication with automatic token refresh
- ⚡ **Async Support**: Full async/await support for high-performance applications
- 🛡️ **Type Safety**: Full type hints and Pydantic models for better IDE support
- 🔁 **Retry Logic**: Automatic retry with exponential backoff for network failures

## Installation

```bash
pip install hexabase-ai
```

## Quick Start

```python
import asyncio
from hexabase_ai import HexabaseClient

async def main():
    # Initialize client
    async with HexabaseClient(api_key="your-api-key") as client:
        # Deploy a function
        deployment = await client.deploy_function(
            name="hello-world",
            code="""
def handler(event, context):
    name = event.get('name', 'World')
    return {
        'statusCode': 200,
        'body': f'Hello, {name}!'
    }
""",
            runtime="python3.9"
        )
        
        # Execute the function
        result = await client.execute_function(
            function_id=deployment.function_id,
            payload={"name": "Hexabase"}
        )
        
        print(result.result)  # {'statusCode': 200, 'body': 'Hello, Hexabase!'}

if __name__ == "__main__":
    asyncio.run(main())
```

## Authentication

The SDK supports multiple authentication methods:

### API Key (Recommended)

```python
# Pass directly
client = HexabaseClient(api_key="your-api-key")

# Or use environment variable
# export HEXABASE_API_KEY=your-api-key
client = HexabaseClient()
```

## Auto-Cleanup

Functions can be automatically cleaned up based on various policies:

```python
from hexabase_ai import CleanupPolicy

# Define cleanup policy
cleanup_policy = CleanupPolicy(
    ttl_hours=24,           # Delete after 24 hours
    max_executions=1000,    # Delete after 1000 executions
    idle_hours=6           # Delete if idle for 6 hours
)

# Deploy with auto-cleanup
deployment = await client.deploy_function(
    name="auto-cleanup-function",
    code="def handler(event, context): return {'status': 'ok'}",
    runtime="python3.9",
    auto_cleanup=cleanup_policy
)
```

## Async Execution

For long-running functions, use asynchronous execution:

```python
# Start async execution
execution = await client.execute_function(
    function_id=function_id,
    payload={"task": "process_large_dataset"},
    async_execution=True
)

# Poll for completion
while execution.status in ["pending", "running"]:
    await asyncio.sleep(5)
    execution = await client.get_execution_status(execution.execution_id)

print(f"Final status: {execution.status}")
```

## Function Configuration

### Supported Runtimes

- `python3.8`
- `python3.9`
- `python3.10`
- `python3.11`
- `nodejs16`
- `nodejs18`

### Resource Configuration

```python
deployment = await client.deploy_function(
    name="resource-intensive-function",
    code=function_code,
    runtime="python3.9",
    memory_mb=512,          # 128MB to 3008MB
    timeout_seconds=300,    # 1 to 900 seconds
    environment={
        "ENV_VAR": "value"
    },
    dependencies=[
        "numpy==1.24.0",
        "pandas==2.0.0"
    ]
)
```

## Error Handling

The SDK provides specific exception types for different error scenarios:

```python
from hexabase_ai import (
    AuthenticationError,
    FunctionNotFoundError,
    FunctionExecutionError,
    ValidationError,
    NetworkError
)

try:
    result = await client.execute_function(function_id, payload)
except FunctionNotFoundError:
    print("Function does not exist")
except FunctionExecutionError as e:
    print(f"Execution failed: {e.details}")
except NetworkError:
    print("Network error occurred")
```

## Advanced Usage

### Deploy from File

```python
deployment = await client.deploy_function(
    name="my-function",
    file_path="./functions/processor.py",
    runtime="python3.9",
    handler="processor.main"
)
```

### List Functions

```python
functions = await client.list_functions(
    limit=50,
    name_filter="data-"
)

for func in functions:
    print(f"{func.name}: {func.status} (executions: {func.execution_count})")
```

### Manual Cleanup

```python
# Delete a specific function
await client.delete_function(function_id)
```

## Development

### Running Tests

```bash
# Install dev dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Run with coverage
pytest --cov=hexabase_ai
```

### Type Checking

```bash
mypy hexabase_ai
```

### Code Formatting

```bash
# Format code
black hexabase_ai tests

# Sort imports
isort hexabase_ai tests
```

## License

MIT License - see LICENSE file for details.