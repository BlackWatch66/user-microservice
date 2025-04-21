.PHONY: build run clean proto docker-build docker-up docker-down test

# Default target
all: build

# Build application
build:
	go build -o bin/user-service ./cmd/server

# Run application
run: build
	./bin/user-service

# Clean build artifacts
clean:
	rm -rf bin/

# Generate Protocol Buffers and gRPC code
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/grpc/proto/user.proto

# Build Docker images
docker-build:
	docker-compose build

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# Run tests using Docker
docker-test:
	docker-compose run --rm user-service go test ./...

# View service logs
logs:
	docker-compose logs -f

# Run tests
test:
	go test ./... 