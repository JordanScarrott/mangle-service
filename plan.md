# Mangle-Service Development Plan

This document outlines the development plan for the `mangle-service`, a log analysis tool for microservices based on Google's Mangle library and a Ports and Adapters architecture.

## 1. Project Setup & Foundational Work

This phase covers the initial setup of the Go project and its basic structure.

*   **[x] Ticket 1: Initialize Go Project & Directory Structure**
    *   **Description:** Create a new Go module for the `mangle-service`. Set up the directory structure to align with the Ports and Adapters architecture: `/cmd/mangle-service`, `/internal/core`, `/internal/adapters`, `/internal/ports`.

*   **[x] Ticket 2: Implement Basic Web Server**
    *   **Description:** Set up a simple HTTP server using `net/http` or a lightweight framework like Gin. This will serve as the foundation for the API endpoints. Create a basic health check endpoint (e.g., `/healthz`).

*   **[x] Ticket 3: Integrate Mangle Library**
    *   **Description:** Add the Mangle library (`github.com/google/mangle`) as a dependency to the project. Create a small proof-of-concept within a temporary main function to ensure it integrates correctly.

## 2. Application Core & Port Definition

This phase focuses on defining the core logic's boundaries through ports.

*   **[x] Ticket 4: Define Log Data Port**
    *   **Description:** In `/internal/ports`, define a `LogDataPort` interface. This interface will abstract the data source and specify methods required by the core application, such as `FetchLogs(queryCriteria map[string]string) ([]Fact, error)`. The `Fact` type will be defined in the core domain.

*   **[x] Ticket 5: Define Core Log Service**
    *   **Description:** In `/internal/core/services`, create a `LogService` that uses the `LogDataPort`. This service will orchestrate the fetching of logs and will be responsible for the primary business logic, independent of external systems.

## 3. Infrastructure: Elasticsearch Adapter

This phase involves implementing the concrete adapter for fetching logs from Elasticsearch.

*   **[x] Ticket 6: Implement Elasticsearch Adapter Stub**
    *   **Description:** In `/internal/adapters/elasticsearch`, create an `ElasticsearchAdapter` struct that implements the `LogDataPort` interface. Initially, this can have stubbed-out methods that return dummy data.

*   **[x] Ticket 7: Implement Elasticsearch Connection Logic**
    *   **Description:** Add logic to the `ElasticsearchAdapter` to connect to an Elasticsearch instance. Connection details should be configurable via environment variables.

*   **[x] Ticket 8: Implement Log Fetching Logic**
    *   **Description:** Implement the `FetchLogs` method in the `ElasticsearchAdapter`. This method will construct and execute a query against an Elasticsearch instance based on the provided criteria.

*   **Ticket 9: Implement Data Transformation to Mangle Facts**
    *   **Description:** Within the `ElasticsearchAdapter`, add the logic to transform the raw JSON log documents received from Elasticsearch into the `Fact` format required by the Mangle library.

## 4. Defining Service Relationships

This phase is about creating a way to define and load the relationships between microservices.

*   **Ticket 10: Design Service Relationship Configuration Format**
    *   **Description:** Define a YAML or JSON schema for specifying microservice relationships. This configuration will define how services communicate, e.g., `service-a -> service-b`.

*   **Ticket 11: Implement Configuration Loader**
    *   **Description:** In `/internal/core/services`, create a `RelationshipService` responsible for loading the service relationship definitions from a file.

*   **Ticket 12: Generate Mangle Rules from Relationships**
    *   **Description:** Add logic to the `RelationshipService` to transform the loaded relationship definitions into a set of Mangle rules that can be used for graph-based queries.

## 5. Querying and Execution

This phase connects the API to the core logic to enable end-to-end querying.

*   **Ticket 13: Create API Endpoint for Mangle Queries**
    *   **Description:** In the web server component (`/cmd/mangle-service`), create a new API endpoint (e.g., `POST /query`) that accepts a text-based Mangle query in the request body.

*   **Ticket 14: Implement Query Execution Logic**
    *   **Description:** In the core `LogService`, implement a method that takes an incoming Mangle query. This method will use the `LogDataPort` to fetch log facts and the `RelationshipService` to get relationship rules.

*   **Ticket 15: Execute Query with Mangle Engine**
    *   **Description:** Use the Mangle engine to execute the query against the combined set of log facts and relationship rules.

*   **Ticket 16: Format and Return Query Results**
    *   **Description:** Format the results from the Mangle engine into a user-friendly JSON structure and return them from the `/query` API endpoint.

## 6. Containerization & Deployment

This phase prepares the service for easy deployment and local development.

*   **Ticket 17: Write Dockerfile**
    *   **Description:** Create a multi-stage `Dockerfile` to build and containerize the `mangle-service`. The final image should be lightweight and include only the compiled binary and necessary assets.

*   **Ticket 18: Create Docker Compose for Local Development**
    *   **Description:** Create a `docker-compose.yml` file. This file will define the `mangle-service` and a mock/test Elasticsearch instance, allowing developers to run the entire stack locally with a single command.
