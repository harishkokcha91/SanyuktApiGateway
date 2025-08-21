# Makefile for managing all microservices and PostgreSQL

# Run all services
run-all:
	docker-compose up -d
# Stop all running containers
stop-all:
	docker-compose stop
# Build all services
build-all:
	docker-compose build
# Rebuild all services (force rebuild)
rebuild-all:
	docker-compose build --no-cache
# View logs for all services
logs:
	docker-compose logs -f
# Bring everything down (stop + remove containers, networks, volumes)
down:
	docker-compose down -v
# Restart all services
restart:
	docker-compose down -v && docker-compose up -d
# Show status of all containers
status:
	docker-compose ps
.PHONY: run-all stop-all build-all rebuild-all logs down restart status \
        run-apigateway run-auth run-brillientstudent run-business \
        run-DBService run-imageUpload run-event run-user run-userProfile
