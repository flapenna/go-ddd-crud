# Go DDD CRUD - Backend Showcase

This project is a backend service built with Golang, implementing Domain-Driven Design (DDD) principles and utilizing gRPC for communication. It serves as a simple **CRUD** (Create, Read, Update, Delete) service for user management, with an emphasis on clean architecture, scalability, and maintainability.

## Project Structure

```plaintext
.
├── build                   # Directory for build-related Docker files
├── cmd
│   └── server
│       └── main.go         # The entry point of the application; contains the main function
├── config
│   └── config.go           # Configuration definitions and loading logic for the application
├── internal                # Private application and library code
│   ├── domain              # Domain layer, defining the core business logic and entities
│   │   └── user            # User-related domain logic, entities, and business rules
│   ├── infrastructure      # Infrastructure layer, containing implementations for external services and data access
│   │   ├── kafka           # Kafka-related infrastructure code (event producer)
│   │   └── mongodb         # MongoDB-related infrastructure code, including repository implementations
│   ├── interfaces
│   │   └── grpc            # gRPC server implementations and definitions
│   └── mocks               # Mock implementations for testing purposes
├── pb                      # Protocol Buffer (protobuf) generated code
├── pkg
│   └── pb                  # Protocol Buffer (protobuf) definitions
├── scripts                 # Scripts for automation, such as database migrations and setup
└── test
    └── integration         # Integration tests to ensure components work together as expected
```

The `internal/domain/user` folder contains all interfaces and business logic for the user service.

## Requirements

- Docker

## Running the Application

Ensure you have Docker installed and running on your system.

Start the application along with all required components (Kafka, MongoDB) using Docker:

```bash
./scripts/docker-start.sh
```

Stop the application:

```bash
./scripts/docker-stop.sh
```

## Port Bindings

- **gRPC**: 9090
- **HTTP**: 8080
- **Kafka**: 9093
- **MongoDB**: 27017
- **Kafdrop**: 9000

## Makefile

A `Makefile` is provided for managing common build and development tasks:

```bash
- buf-migrate       # Migrates Buf configuration to the latest version using bufbuild/buf Docker image
- proto             # Updates dependencies and generates protobuf files using bufbuild/buf Docker image
- clean             # Removes the compiled binary `bin/go-ddd-crud`
- build             # Cleans previous build and compiles the Go application `bin/go-ddd-crud`
- run               # Builds and runs the Go application
- mocks             # Generates mock implementations for testing using vektra/mockery Docker image
- test              # Runs all tests (unit and integration) and generates a coverage report
- unit              # Runs only unit tests and generates a coverage report
- integration       # Runs only integration tests and generates a coverage report
- coverage          # Generates an HTML coverage report from the `coverage.out` file
```

## gRPC and Protobuf

There are 3 proto files under the `pb` folder:

- `user_service.proto`: Defines services and messages related to user management.
- `user_event.proto`: Defines messages for user-related events.
- `health_service.proto`: Defines services and messages for health checks.

We use [Buf](https://buf.build/docs/ecosystem/cli-overview) for compiling Protobuf files. Buf simplifies the process with features like proto linting, dependency management, and CI/CD integration.

### REST API Support

Although the service primarily exposes a gRPC API on port 9090, an HTTP gateway is provided at port 8080 using [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway). This translates RESTful requests into gRPC calls, enabling easier testing of the API.

### Update User

The **Update User** endpoint uses the **PUT** method. The request must contain all user fields: `first_name`, `last_name`, `country`, `email`, and `nickname`.

### List Users

The **ListUsers** endpoint supports optional filter parameters for `first_name`, `last_name`, `country`, and `nickname`, as well as pagination parameters `page` and `page_size`. The server defaults to `page=0` and `page_size=10` if not provided.

## MongoDB Change Streams

To showcase event-driven design, MongoDB Change Streams are implemented to watch for changes to user entities. This is a basic implementation without horizontal scaling or resume token support, but it demonstrates how to notify external services when user data changes.

For production systems, consider more robust solutions like the **Outbox Pattern**, **Change Data Capture (CDC)**, or **Event Sourcing** to ensure atomicity between database writes and event publishing.

## Testing

The project contains both unit and integration tests to ensure the correctness of the codebase. The tests can be run using the `Makefile`.

- **Unit tests** cover individual components and business logic.
- **Integration tests** verify that different parts of the system work together.

Test files are located under the `test/integration` folder, which contains gRPC client tests for verifying API interactions.

## Logging

The application uses `logrus` for structured logging in JSON format.

- **Info**: Logs key events like starting/stopping services.
- **Warning**: Logs recoverable errors or warnings.
- **Error**: Logs critical errors.
- **Debug**: Logs detailed development information (disabled in production).

## MongoDB Indexing

Although no explicit indexes are created in the application, consider indexing fields like `email` and `nickname` in MongoDB to improve query performance, as these fields are stable and frequently queried.

## Caching

Caching is not implemented, but Redis would be a good option for reducing database load given that user data is relatively stable. The **look-aside caching** pattern could significantly improve performance in production environments.

## Documentation

- [Postman Collection](./docs/GO-DDD_HTTP.postman_collection.json)
- [Swagger](./docs/GO-DDD-user_service.swagger.json)

---

This showcase aims to demonstrate how to build a maintainable, scalable, and testable backend service in Go using DDD and gRPC. It includes the core functionalities needed for user management and can be extended with additional features like caching, more robust event sourcing, or external integrations.

---