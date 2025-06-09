"""Example demonstrating asynchronous function execution."""

import asyncio
from hexabase_ai import HexabaseClient


async def main():
    # Initialize client
    async with HexabaseClient(api_key="your-api-key") as client:
        # Deploy a long-running function
        print("Deploying long-running function...")
        deployment = await client.deploy_function(
            name="long-task",
            code="""
import time
import json

def handler(event, context):
    duration = event.get('duration', 10)
    
    # Simulate long-running task
    time.sleep(duration)
    
    return {
        'statusCode': 200,
        'body': json.dumps({
            'message': f'Task completed after {duration} seconds',
            'input': event
        })
    }
""",
            runtime="python3.9",
            timeout_seconds=300  # 5 minute timeout
        )
        
        print(f"Function deployed: {deployment.function_id}")
        
        # Execute function asynchronously
        print("\nStarting async execution...")
        execution = await client.execute_function(
            function_id=deployment.function_id,
            payload={"duration": 5, "task": "process_data"},
            async_execution=True
        )
        
        print(f"Execution started: {execution.execution_id}")
        print(f"Initial status: {execution.status}")
        
        # Poll for completion
        print("\nPolling for completion...")
        while execution.status in ["pending", "running"]:
            await asyncio.sleep(2)
            execution = await client.get_execution_status(execution.execution_id)
            print(f"  Status: {execution.status}")
        
        # Final result
        print(f"\nFinal status: {execution.status}")
        if execution.status == "completed":
            print(f"Result: {execution.result}")
        elif execution.error:
            print(f"Error: {execution.error}")
        
        print(f"Duration: {execution.duration_ms}ms")
        print(f"Billed duration: {execution.billed_duration_ms}ms")


if __name__ == "__main__":
    asyncio.run(main())