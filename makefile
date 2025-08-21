# Makefile for managing microservices and PostgreSQL

# Docker-based operations
run-all-docker:
	docker-compose up -d

stop-all:
	docker-compose stop

build-all:
	docker-compose build

rebuild-all:
	docker-compose build --no-cache

logs:
	docker-compose logs -f

down:
	docker-compose down -v

restart:
	docker-compose down -v && docker-compose up -d

status:
	docker-compose ps

# Local Go-based execution
run-all-local: run-db-local
	$(MAKE) run-apigateway &
	$(MAKE) run-auth &
	$(MAKE) run-brillientstudent &
	$(MAKE) run-business &
	$(MAKE) run-DBService &
	$(MAKE) run-imageUpload &
	$(MAKE) run-event &
	$(MAKE) run-user &
	$(MAKE) run-userProfile &
	wait

# Individual microservice runners
run-apigateway:
	cd APiGateway && go run main.go

run-auth:
	cd AuthService && go run main.go

run-brillientstudent:
	cd BrillientStudentAchievement && go run main.go

run-business:
	cd BusinessMicroservice && go run main.go

run-DBService:
	cd DBService && go run main.go

run-imageUpload:
	cd ImageUploadService && go run main.go

run-event:
	cd NamdevEvents && go run main.go

run-user:
	cd UserMicroservice && go run main.go

run-userProfile:
	cd UserProfileMicroservice && go run main.go


# Start only PostgreSQL container for local development
run-db-local:
	docker-compose -f docker-compose.yml up -d postgres

# Help command
help:
    @echo "Usage: make [target]"
	@echo ""
    @echo "Docker-based targets:"
	@echo "  run-all-docker     Run all services via Docker Compose"
	@echo "  stop-all           Stop all running containers"
	@echo "  build-all          Build all services"
	@echo "  rebuild-all        Rebuild all services (no cache)"
	@echo "  logs               View logs for all services"
	@echo "  down               Stop and remove containers, networks, volumes"
	@echo "  restart            Restart all services"
	@echo "  status             Show container status"
	@echo ""
	@echo "Local Go-based targets:"
	@echo "  run-all-local      Run all services locally via go run"
	@echo "  run-<service>      Run individual service (e.g., run-auth)"
	@echo ""
	@echo "Use 'make help' to see this message again."

# Declare phony targets
.PHONY: run-all-docker stop-all build-all rebuild-all logs down restart status \
		run-all-local run-apigateway run-auth run-brillientstudent run-business \
		run-DBService run-imageUpload run-event run-user run-userProfile help
