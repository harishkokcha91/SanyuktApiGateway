.PHONY: run-apigateway run-auth run-brillientstudent run-business run-DBSerivce run-imageUpload run-event run-user run-userProfile run-all

# Run the apigateway microservice
run-apigateway:
	cd APiGateway && go run main.go

# Run the auth microservice
run-auth:
	cd AuthService && go run main.go

# Run the Brillient Student microservice
run-brillientstudent:
	cd BrillientStudentAchievement && go run main.go

# Run the Business microservice
run-business:
	cd BusinessMicroservice && go run main.go

# Run the Business microservice
run-DBSerivce:
	cd DBService && go run main.go

# Run the imageUpload microservice
run-imageUpload:
	cd ImageUploadService && go run main.go

# Run the Event microservice
run-event:
	cd NamdevEvents && go run main.go

# Run the User microservice
run-user:
	cd UserMicroservice && go run main.go

# Run the Profile microservice
run-userProfile:
	cd UserProfileMicroservice && go run main.go

# Run all microservices concurrently
run-all:
	make run-apigateway &
	make run-auth &
	make run-brillientstudent &
	make run-business &
	make run-imageUpload &
	make run-event &
	make run-user &
	make run-userProfile &
	wait
