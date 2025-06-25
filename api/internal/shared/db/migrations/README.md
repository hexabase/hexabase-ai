# Database Migrations

This document explains how to manage database schema changes using `golang-migrate`. All schema changes must be done following the process described here.

## 1. Overview

To ensure database schema consistency and traceability across all environments, we use `golang-migrate` to manage schema changes through version-controlled SQL files.

**Key Principles:**

- **GORM `AutoMigrate` is Deprecated**: The use of `gorm.AutoMigrate` has been completely removed. Modifying Go models will no longer alter the database schema.
- **SQL-First Approach**: All schema changes must be explicitly defined in `.sql` migration files.
- **Version Control**: Every schema change is versioned and tracked in Git.
- **Reversibility**: Each migration must have a corresponding "down" migration to allow for safe rollbacks.

This migration strategy was decided upon in [ADR-012: Unification of Database Migration](../../../docs/adr/011-unify-database-migration.md).

## 2. How Migrations Are Executed

Migrations are automatically applied when the API server is started in a development environment. This is handled by the `migrate` service in Docker Compose, which is triggered by the following commands:

- `make dev-api`
- `make debug-api`

The service checks for any new migration files in this directory and applies the `up` migrations sequentially to the local PostgreSQL database.

## 3. Developer Workflow

Follow these steps to create and apply a new database schema change.

### Step 1: Managing and `migrate` Tool

The `golang-migrate` CLI is managed as a project dependency using the `tool` directive in our `go.mod` file (a feature available since Go 1.24). This ensures all developers use the same version of the tool.

The primary way to run this tool is via the `go tool` command, which requires no extra setup.

For convenience, if you wish to use the `migrate` command directly in your terminal, you can install it into your local Go environment by running:
```bash
go install tool
```
This will place the `migrate` binary in your `$GOPATH/bin` directory. For the `migrate` command to be accessible directly, you must have this directory in your system's `PATH` environment variable.

### Step 2: Create a New Migration File

To generate a new pair of migration files, you should use the `go tool migrate` command. This is the recommended approach as it works regardless of your shell's `PATH` configuration.

```bash
go tool migrate create -ext sql -dir api/internal/shared/db/migrations -format "20060102150405" your_migration_name_in_snake_case
```

**Example:**
```bash
go tool migrate create -ext sql -dir api/internal/shared/db/migrations -format "20060102150405" add_api_key_to_users
```

> **Note:** If you have successfully run `go install tool` and configured environment variables appropriately, you can use the shorter `migrate` command instead of `go tool migrate`.

This will create two files in `api/internal/shared/db/migrations/`:
- `YYYYMMDDHHMMSS_add_api_key_to_users.up.sql`
- `YYYYMMDDHHMMSS_add_api_key_to_users.down.sql`

> **Tip:** For convenience, you can also use the `make migrate-create` command, which will prompt you for a migration name and run the `go tool migrate` command for you.

### Step 3: Write the Migration SQL

Edit the generated files with your SQL statements.

- **`*.up.sql`**: Contains the SQL to apply your changes.
- **`*.down.sql`**: Contains the SQL to revert the changes from the `up` file.

**Example `up.sql`:**
```sql
-- Add an api_key column to the users table
ALTER TABLE "users" ADD COLUMN "api_key" VARCHAR(255);
```

**Example `down.sql`:**
```sql
-- Remove the api_key column from the users table
ALTER TABLE "users" DROP COLUMN "api_key";
```

**Best Practices:**
- **Idempotency**: Whenever possible, write statements that can be run multiple times without causing errors (e.g., `CREATE TABLE IF NOT EXISTS`).
- **Transactions**: Each migration file is executed within a single transaction. Do not include `BEGIN;` or `COMMIT;`.
- **Reversibility**: Ensure your `down` migration correctly and completely reverts the `up` migration.

### Step 4: Test Your Migration Locally

1.  Apply your migration by running the API server:
    ```bash
    make dev-api
    ```
    Observe the logs to confirm that your migration was applied successfully.

2.  (Optional but recommended) Test the `down` migration by running the `migrate` service directly:
    ```bash
    # Roll back the last migration
    docker compose -f docker-compose.yml --profile tools run --rm migrate down 1

    # Re-apply the migration
    docker compose -f docker-compose.yml --profile tools run --rm migrate up
    ```
    This ensures that both directions of your migration work as expected.

## 4. Troubleshooting

### Dirty Migration State

If a migration script fails, the database can be left in a "dirty" state, which prevents new migrations from running.

**To fix a dirty state (for local development):**
1.  Identify the migration version that failed from the logs.
2.  Fix the SQL in the problematic migration file.
3.  Manually revert any partial changes from the database if necessary.
4.  Force the `migrate` tool to set the version to the one *before* the failed one.
    ```bash
    # Example: if version 4 is dirty, force it back to version 3
    docker compose -f docker-compose.yml --profile tools run --rm migrate force 3
    ```
5.  Run the migration again with `make dev-api`.

**Warning**: The `force` command is a destructive operation intended for local development environments only. It should **never** be used in staging or production. 