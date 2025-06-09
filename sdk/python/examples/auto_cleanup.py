"""Example demonstrating auto-cleanup functionality."""

import asyncio
from hexabase_ai import HexabaseClient, CleanupPolicy


async def main():
    # Initialize client
    client = HexabaseClient(api_key="your-api-key")
    await client.authenticate()
    
    try:
        # Deploy a function with auto-cleanup policy
        print("Deploying function with auto-cleanup...")
        
        cleanup_policy = CleanupPolicy(
            ttl_hours=24,           # Delete after 24 hours
            max_executions=100,     # Delete after 100 executions
            idle_hours=6           # Delete if idle for 6 hours
        )
        
        deployment = await client.deploy_function(
            name="auto-cleanup-demo",
            code="""
import json
import datetime

def handler(event, context):
    return {
        'statusCode': 200,
        'body': json.dumps({
            'message': 'This function will auto-cleanup!',
            'timestamp': datetime.datetime.utcnow().isoformat()
        })
    }
""",
            runtime="python3.9",
            dependencies=["python-dateutil"],
            auto_cleanup=cleanup_policy,
            cleanup_interval=300  # Check every 5 minutes
        )
        
        print(f"Function deployed: {deployment.function_id}")
        print("Auto-cleanup policy applied:")
        print(f"  - TTL: 24 hours")
        print(f"  - Max executions: 100")
        print(f"  - Idle timeout: 6 hours")
        
        # Execute the function a few times
        print("\nExecuting function multiple times...")
        for i in range(3):
            result = await client.execute_function(
                function_id=deployment.function_id,
                payload={"iteration": i}
            )
            print(f"  Execution {i+1}: {result.status}")
            await asyncio.sleep(1)
        
        # The function will be automatically cleaned up based on the policy
        print("\nFunction will be automatically cleaned up based on the policy.")
        print("No manual cleanup required!")
        
    finally:
        # Cleanup client resources
        await client._cleanup()


if __name__ == "__main__":
    asyncio.run(main())