#!/bin/bash
# Development runner script for Hexabase API

# Load environment variables from .env file
if [ -f .env ]; then
    echo "Loading environment variables from .env..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Warning: .env file not found!"
fi

# Show database connection info
echo "Connecting to database:"
echo "  Host: ${DATABASE_HOST}:${DATABASE_PORT}"
echo "  Database: ${DATABASE_DBNAME}"
echo "  User: ${DATABASE_USER}"

# Run the API
echo "Starting API server..."
go run cmd/api/main.go