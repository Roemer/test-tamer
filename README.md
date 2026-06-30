# Test Tamer

## Requirements

For Postgres Backend: Postgres 18

## Configuration

### Environment variables
| Variable | Default Value | Description |
| --- | --- | --- |
| `TEST_TAMER_DB_TYPE` | `postgres` | Backend database type to use. Can be `postgres`, `file` or `memory`. |
| `TEST_TAMER_POSTGRES_USER` | `test_tamer` | PostgreSQL username for authentication. |
| `TEST_TAMER_POSTGRES_PASSWORD` | No default set | PostgreSQL password for authentication. |
| `TEST_TAMER_POSTGRES_DB` | `test_tamer` | PostgreSQL database name. |
| `TEST_TAMER_POSTGRES_HOST` | `postgres` | PostgreSQL server host. |
| `TEST_TAMER_POSTGRES_PORT` | `5432` | PostgreSQL server port. |
| `TEST_TAMER_POSTGRES_SSLMODE` | `(empty)` | PostgreSQL SSL mode (`disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`). |
