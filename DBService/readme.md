# Database Service - README

## Overview
This service is responsible for creating and managing the database for your microservices. It sets up the necessary tables and handles initial database configurations.

## Prerequisites
Ensure you have the following installed on your system:
- Go (version 1.19 or later)
- Docker (if using Docker for database setup)
- PostgreSQL or MySQL (depending on your configuration)

## Installation and Setup

### 1. Initialize the Go Module
Run the following command inside the project directory:
```sh
go mod init db-service
```

### 2. Install Dependencies
```sh
go mod tidy
```

### 3. Configure Environment Variables
Create a `.env` file in the root directory and set the required database credentials:
```
DB_HOST=localhost
DB_PORT=6501
DB_USER=postgres
DB_PASSWORD=admin
DB_NAME=sanyukt
```

### 4. Run the Service
#### Without Docker
```sh
go run main.go
```

#### With Docker
If you have a `Dockerfile` and `docker-compose.yml`, run:
```sh
docker-compose up --build
```

## Database Migrations

The migration CLI is located at `./cmd/migrate`. All commands should be run from the `DBService` directory.

### Run all pending migrations (default)
```sh
go run ./cmd/migrate migrate
```

### Check migration status
```sh
go run ./cmd/migrate migrate -status
```

### Rollback the last migration
```sh
go run ./cmd/migrate migrate -down
```

### Initialize database (create if not exists) then run all migrations
```sh
go run ./cmd/migrate init
```

### List all tables in the database
```sh
go run ./cmd/migrate tables
```

### Create new migration files (manual step - creates template names)
```sh
go run ./cmd/migrate migrate -create <migration_name>
```

### Environment Variables for Migrations
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `6501` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `admin` |
| `DB_NAME` | Database name | `sanyukt` |
| `MIGRATIONS_PATH` | Path to migrations directory (optional, auto-detected) | Auto-detected |

### Migration Path Auto-Detection
The migrations path is automatically detected in these locations (in order):
1. `MIGRATIONS_PATH` environment variable (for Docker, AWS, CI/CD)
2. `../migrations` (from DBService/)
3. `../../migrations` (from DBService/cmd/migrate/)
4. `migrations` (from project root)
5. `/app/migrations` (Docker default)
6. `/migrations` (alternative Docker)

### Example: Run migrations in Docker
```sh
docker run --rm \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=admin \
  -e DB_NAME=sanyukt \
  -e MIGRATIONS_PATH=/app/migrations \
  -v $(pwd)/migrations:/app/migrations \
  your-db-service-image \
  go run ./cmd/migrate migrate
```

## API Endpoints
- `GET /health` - Check service health
- `POST /init-db` - Initialize database schema

## Contributing
Feel free to fork this repository and contribute. Ensure you follow best practices for Go and database management.

## License
This project is open-source and available under the MIT License.