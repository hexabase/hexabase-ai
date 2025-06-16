#!/bin/bash

# Script to run database migrations

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values (using correct env var names from .env)
DB_HOST=${DATABASE_HOST:-localhost}
DB_PORT=${DATABASE_PORT:-5432}
DB_USER=${DATABASE_USER:-postgres}
DB_PASSWORD=${DATABASE_PASSWORD:-postgres}
DB_NAME=${DATABASE_DBNAME:-hexabase_ai}

echo "Running migrations on database: $DB_NAME@$DB_HOST:$DB_PORT"

# Run all up migrations in order
for migration in api/internal/db/migrations/*.up.sql; do
    echo "Running migration: $(basename $migration)"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration"
done

echo "Migration completed successfully!" 