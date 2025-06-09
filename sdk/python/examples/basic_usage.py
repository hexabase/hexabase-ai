"""Basic usage example for Hexabase AI SDK."""

import asyncio
from hexabase_ai import HexabaseClient, CleanupPolicy


async def main():
    # Initialize client
    async with HexabaseClient(api_key="your-api-key") as client:
        # Deploy a simple function
        print("Deploying function...")
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
            runtime="python3.9",
            handler="handler"
        )
        
        print(f"Function deployed: {deployment.function_id}")
        print(f"Endpoint: {deployment.endpoint}")
        
        # Execute the function
        print("\nExecuting function...")
        result = await client.execute_function(
            function_id=deployment.function_id,
            payload={"name": "Hexabase"}
        )
        
        print(f"Execution ID: {result.execution_id}")
        print(f"Status: {result.status}")
        print(f"Result: {result.result}")
        
        # List functions
        print("\nListing functions...")
        functions = await client.list_functions()
        for func in functions:
            print(f"- {func.name} ({func.function_id}): {func.status}")


if __name__ == "__main__":
    asyncio.run(main())