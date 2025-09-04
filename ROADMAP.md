# ROADMAP.md

This document outlines the future development direction for the mangle-service. These improvements are designed to make the service more robust, maintainable, and easier to operate.

## 1. Containerization

**Status:** Partially implemented.

The project currently has a `Dockerfile`, but it is outdated and references incorrect paths. The next steps are to:

*   **Fix the `Dockerfile`:** Update the `Dockerfile` to correctly build the application from `./cmd/server/main.go`.
*   **Add `docker-compose.yml`:** Create a `docker-compose.yml` file to simplify local development and testing, allowing developers to run the service and any dependencies (like databases or other services) with a single command.

## 2. CI/CD Pipeline

**Status:** Not started.

To ensure code quality and automate deployments, a CI/CD pipeline should be implemented.

*   **GitHub Actions:** Create a basic GitHub Actions workflow that triggers on every push to the `main` branch. This workflow should run the test suite (`go test -v ./...`) to catch any regressions early.

## 3. Configuration Management

**Status:** Not implemented.

Currently, configuration is handled via environment variables, which can be limiting.

*   **Implement Viper:** Integrate a library like [Viper](https://github.com/spf13/viper) to handle configuration. This will allow for more flexible configuration options, such as using configuration files (e.g., `config.yaml`) and environment variables.

## 4. Structured Logging

**Status:** Partially implemented.

The service uses the standard `log/slog` package, but it could be enhanced.

*   **Integrate a dedicated logging library:** While `slog` is a good start, integrating a more feature-rich structured logging library like [zerolog](https://github.com/rs/zerolog) can provide better performance and more flexible output options (e.g., JSON formatting for easy parsing by log aggregators).

## 5. API Documentation

**Status:** Not implemented.

To make the API easy to understand and use, it should be well-documented.

*   **Add Swagger/OpenAPI:** Integrate [Swagger/OpenAPI](https://swagger.io/) to generate interactive API documentation directly from the code. This will make it easier for frontend developers and other consumers of the API to understand the available endpoints and their usage.

## 6. Health Check Endpoint

**Status:** Not implemented.

For monitoring and orchestration purposes, the service needs a health check endpoint.

*   **Implement `/healthz`:** Add a `/healthz` endpoint that returns a `200 OK` status if the service is healthy. This is a standard practice for services running in orchestrated environments like Kubernetes.
