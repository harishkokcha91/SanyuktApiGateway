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
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=your_database
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

### 5. Database Migrations (if applicable)
If you are using GORM or a migration tool, run:
```sh
go run migrate.go
```

## API Endpoints
- `GET /health` - Check service health
- `POST /init-db` - Initialize database schema

## Contributing
Feel free to fork this repository and contribute. Ensure you follow best practices for Go and database management.

## License
This project is open-source and available under the MIT License.


3️⃣ Check the Database Connection
## Run the following command to check if your database is actually running in Docker:
```sh
docker ps
```
## If PostgreSQL is not running, restart it:
```sh
docker-compose up -d
```
## Verify connection from inside the container:
```sh
docker exec -it postgres-db psql -U postgres -d sanyukt
```
## Restart your Docker containers:
```sh
docker-compose down && docker-compose up -d
```