# User Microservice

A Go-based microservice providing user registration, login, profile and address management.

## Features

- User registration and login (JWT authentication)
- User profile management
- User address management
- Support for both HTTP and gRPC interfaces
- MySQL for user data storage
- Redis for JWT token caching

## Tech Stack

- Go 1.23+
- Gin Web Framework
- gRPC
- GORM (MySQL)
- Redis
- Docker & Docker Compose

## Quick Start

### Requirements

- Docker and Docker Compose
- Or locally installed Go 1.23+, MySQL and Redis

### Running with Docker Compose

1. Clone the repository:

```bash
git clone https://github.com/blackwatch66/user-microservice.git
cd user-microservice
```

2. Modify the `.env` file settings (optional):

```
# Modify environment variables as needed
JWT_SECRET=your-own-secret-key
```

3. Start the services:

```bash
docker-compose up -d
```

4. Verify that the service is running properly:

```bash
# Test the HTTP API
curl http://localhost:8080/api/users/login -X POST -d '{"email":"test@example.com","password":"test123"}' -H "Content-Type: application/json"

# The test result should return a JSON response containing a token
```

### Docker Compose Configuration Guide

The project uses Docker Compose to manage all required services. Here's a detailed breakdown:

1. **Environment Variables**: Create or modify the `.env` file in the project root with the following variables:

```
# Database configuration
MYSQL_ROOT_PASSWORD=password
MYSQL_DATABASE=user_db
MYSQL_PORT=3306

# Redis configuration
REDIS_PORT=6379
REDIS_PASSWORD=

# User service configuration
HTTP_PORT=8080
GRPC_PORT=50051
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY_MINUTES=15
```

2. **Docker Compose Commands**:

   - **Start services**: `docker-compose up -d`
   - **Stop services**: `docker-compose down`
   - **View logs**: `docker-compose logs -f user-service`
   - **Rebuild services**: `docker-compose build`
   - **Restart a specific service**: `docker-compose restart user-service`

3. **Service Endpoints**:

   - **HTTP API**: `http://localhost:8080`
   - **gRPC API**: `localhost:50051`
   - **MySQL**: `localhost:3306`
   - **Redis**: `localhost:6379`

4. **Volume Management**:

   The Docker Compose setup creates persistent volumes for MySQL and Redis data:
   - `mysql-data`: Stores database files
   - `redis-data`: Stores Redis data

   To remove volumes when stopping services:
   ```bash
   docker-compose down -v
   ```

5. **Scaling**:

   To run multiple instances of the user service:
   ```bash
   docker-compose up -d --scale user-service=3
   ```
   (Note: This requires additional load balancer configuration)

## API Documentation

### HTTP REST API

#### User Registration

- **URL**: `/api/users/signup`
- **Method**: `POST`
- **Authentication**: Not required
- **Request Body**:
  ```json
  {
    "email": "string",       // Required, valid email
    "password": "string"     // Required, minimum 6 characters
  }
  ```
- **Success Response** (201 Created):
  ```json
  {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-04-21T10:23:37.961Z"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
  - 409 Conflict: Email already exists
  - 500 Internal Server Error: Server error

#### User Login

- **URL**: `/api/users/login`
- **Method**: `POST`
- **Authentication**: Not required
- **Request Body**:
  ```json
  {
    "email": "string",      // Required, valid email
    "password": "string"    // Required
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "token": "jwt_token_string"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
  - 401 Unauthorized: Invalid email or password
  - 500 Internal Server Error: Server error

#### Get User Profile

- **URL**: `/api/users/{id}`
- **Method**: `GET`
- **Authentication**: JWT token required
- **Path Parameters**: 
  - `id`: User ID
- **Success Response** (200 OK):
  ```json
  {
    "id": 1,
    "email": "user@example.com",
    "first_name": "string",
    "last_name": "string",
    "created_at": "2025-04-21T10:23:37.961Z",
    "updated_at": "2025-04-21T10:23:37.961Z",
    "addresses": []  // Address array
  }
  ```
- **Error Responses**:
  - 401 Unauthorized: Not authenticated or invalid token
  - 403 Forbidden: No permission to access this resource
  - 404 Not Found: User doesn't exist
  - 500 Internal Server Error: Server error

#### Update User Profile

- **URL**: `/api/users/{id}`
- **Method**: `PUT`
- **Authentication**: JWT token required
- **Path Parameters**: 
  - `id`: User ID
- **Request Body**:
  ```json
  {
    "first_name": "string",
    "last_name": "string"
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "id": 1,
    "email": "user@example.com",
    "first_name": "updated_first_name",
    "last_name": "updated_last_name",
    "created_at": "2025-04-21T10:23:37.961Z",
    "updated_at": "2025-04-21T10:24:37.961Z"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
  - 401 Unauthorized: Not authenticated or invalid token
  - 403 Forbidden: No permission to access this resource
  - 404 Not Found: User doesn't exist
  - 500 Internal Server Error: Server error

#### Get User Address List

- **URL**: `/api/users/{id}/addresses`
- **Method**: `GET`
- **Authentication**: JWT token required
- **Path Parameters**: 
  - `id`: User ID
- **Success Response** (200 OK):
  ```json
  [
    {
      "id": 1,
      "user_id": 1,
      "street": "123 Main St",
      "city": "Beijing",
      "state": "Beijing",
      "postal_code": "100000",
      "country": "China",
      "is_default": true,
      "created_at": "2025-04-21T10:32:49.971Z",
      "updated_at": "2025-04-21T10:32:49.971Z"
    }
  ]
  ```
- **Error Responses**:
  - 401 Unauthorized: Not authenticated or invalid token
  - 403 Forbidden: No permission to access this resource
  - 500 Internal Server Error: Server error

#### Add User Address

- **URL**: `/api/users/{id}/addresses`
- **Method**: `POST`
- **Authentication**: JWT token required
- **Path Parameters**: 
  - `id`: User ID
- **Request Body**:
  ```json
  {
    "street": "string",     // Required
    "city": "string",       // Required
    "state": "string",      // Optional
    "postal_code": "string", // Required
    "country": "string",    // Required
    "is_default": false     // Optional, default is false
  }
  ```
- **Success Response** (201 Created):
  ```json
  {
    "id": 1,
    "user_id": 1,
    "street": "123 Main St",
    "city": "Beijing",
    "state": "Beijing",
    "postal_code": "100000",
    "country": "China",
    "is_default": true,
    "created_at": "2025-04-21T10:32:49.971Z",
    "updated_at": "2025-04-21T10:32:49.971Z"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
  - 401 Unauthorized: Not authenticated or invalid token
  - 403 Forbidden: No permission to access this resource
  - 500 Internal Server Error: Server error

#### Update User Address

- **URL**: `/api/users/{id}/addresses/{addrId}`
- **Method**: `PUT`
- **Authentication**: JWT token required
- **Path Parameters**: 
  - `id`: User ID
  - `addrId`: Address ID
- **Request Body**:
  ```json
  {
    "street": "string",     // Optional
    "city": "string",       // Optional
    "state": "string",      // Optional
    "postal_code": "string", // Optional
    "country": "string",    // Optional
    "is_default": false     // Optional
  }
  ```
- **Success Response** (200 OK):
  ```json
  {
    "id": 1,
    "user_id": 1,
    "street": "456 New St",
    "city": "Shanghai",
    "state": "Shanghai",
    "postal_code": "200000",
    "country": "China",
    "is_default": true,
    "created_at": "2025-04-21T10:32:49.971Z",
    "updated_at": "2025-04-21T10:35:49.971Z"
  }
  ```
- **Error Responses**:
  - 400 Bad Request: Invalid input
  - 401 Unauthorized: Not authenticated or invalid token
  - 403 Forbidden: No permission to access this resource
  - 404 Not Found: Address doesn't exist or doesn't belong to user
  - 500 Internal Server Error: Server error

#### Delete User Address

- **URL**: `/api/users/{id}/addresses/{addrId}`
- **Method**: `DELETE`
- **Authentication**: JWT token required
- **Path Parameters**: 
  - `id`: User ID
  - `addrId`: Address ID
- **Success Response** (204 No Content)
- **Error Responses**:
  - 401 Unauthorized: Not authenticated or invalid token
  - 403 Forbidden: No permission to access this resource
  - 404 Not Found: Address doesn't exist or doesn't belong to user
  - 500 Internal Server Error: Server error

### gRPC API

The user service provides gRPC interfaces on port 50051.

#### CreateUser

- **Request** (CreateUserRequest):
  ```protobuf
  message CreateUserRequest {
    string email = 1;      // User email
    string password = 2;   // Raw password
  }
  ```
- **Response** (CreateUserResponse):
  ```protobuf
  message CreateUserResponse {
    uint64 user_id = 1;    // Created user ID
    string email = 2;      // User email
  }
  ```

#### ValidateToken

- **Request** (ValidateTokenRequest):
  ```protobuf
  message ValidateTokenRequest {
    string token = 1;      // JWT token
  }
  ```
- **Response** (ValidateTokenResponse):
  ```protobuf
  message ValidateTokenResponse {
    bool valid = 1;        // Whether the token is valid
    uint64 user_id = 2;    // If valid, returns user ID
    string email = 3;      // If valid, returns user email
  }
  ```

## Development Guide

### Local Development

1. Install dependencies:

```bash
go mod download
```

2. Generate gRPC code (requires protoc and related plugins):

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/grpc/proto/user.proto
```

3. Set environment variables:

```bash
export DATABASE_URL="root:password@tcp(localhost:3306)/user_db?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR="localhost:6379"
export JWT_SECRET="your-jwt-secret-key"
export HTTP_PORT="8080"
export GRPC_PORT="50051"
```

4. Run the service:

```bash
go run cmd/server/main.go
```

## Docker Build

Build the user service image individually:

```bash
docker build -t user-microservice .
```

## License

MIT


