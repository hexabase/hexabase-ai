#!/bin/bash

# Run tests for Python SDK with coverage

echo "Setting up Python environment..."

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install package in development mode
echo "Installing package and dependencies..."
pip install -e ".[dev]"

# Run tests with coverage
echo "Running tests with coverage..."
pytest -v --cov=hexabase_ai --cov-report=term-missing --cov-report=html

# Run type checking
echo "Running type checking..."
mypy hexabase_ai

# Run linting
echo "Running linting..."
ruff check hexabase_ai tests

# Check code formatting
echo "Checking code formatting..."
black --check hexabase_ai tests
isort --check-only hexabase_ai tests

echo "Test run complete!"
echo "Coverage report available at: htmlcov/index.html"