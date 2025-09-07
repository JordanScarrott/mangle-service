# Mangle Service for Microservice Log Analysis

Mangle Service is a wrapper around Google's Mangle, designed with a Ports and Adapters architecture. It runs deductive queries against microservice logs stored in Elasticsearch, enabling powerful analysis of distributed system behavior.

## Core Features

*   **Unified Log Analysis**: Query disparate log data from Elasticsearch as a unified source.
*   **Relationship-Based Queries**: Define and reason over complex microservice relationships to trace transactions and dependencies.
*   **Extensible by Design**: Built with a Ports and Adapters architecture for easy integration with other data sources and tools.

## Installation and Setup

### Prerequisites

*   **Go**: Version 1.24.3 or later.
*   **Docker**: Latest version.
*   **Docker Compose**: Latest version.

### Local Build

1.  **Clone the repository**:
    ```bash
    git clone <repository-url>
    cd mangle-service
    ```

2.  **Build the application**:
    ```bash
    go build -o mangle-service ./cmd/mangle-service/main.go
    ```

### Docker Compose

For a containerized setup, use Docker Compose. This will spin up the Mangle Service and a compatible Elasticsearch instance.

1.  **Run with Docker Compose**:
    ```bash
    docker-compose up --build
    ```

## Configuration

The service is configured using environment variables.

| Variable                  | Description                                                                                             | Example                               |
| ------------------------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| `MANGLE_SERVICE_PORT`     | The port on which the service will run. Internally maps to the `PORT` variable.                         | `8080`                                |
| `ELASTICSEARCH_URL`       | The full URL for the Elasticsearch instance. Internally maps to the `ELASTICSEARCH_ADDRESS` variable.   | `http://localhost:9200`               |
| `ELASTICSEARCH_INDEX`     | The name of the Elasticsearch index containing the logs. (Note: Currently hardcoded to `logs`)            | `logs`                                |
| `RELATIONSHIPS_CONFIG_PATH` | The file path to the service relationship definitions.                                                  | `config/relationships.yml`            |

## Quick Start Guide: Your First Query

This guide will walk you through defining service relationships and running a simple query.

### Step 1: Define Your Service Relationships

Create a `relationships.yml` file to inform Mangle how your services interact. For example, to define that an `api-gateway` calls an `order-service`, create the following file in `config/relationships.yml`:

```yaml
relationships:
  - name: "api_gateway_to_order_service"
    from: "api-gateway"
    to: "order-service"
    attributes:
      - key: "traceId"
        value: "log.traceId"
```

### Step 2: Write a Mangle Query

Queries are sent as plain text in an API request. Let's write a query to find all error logs from the `order-service`.

```
log.level="error", log.service="order-service"
```

### Step 3: Execute the Query

Use `curl` to send the query to the running service.

```bash
curl -X POST http://localhost:8080/query \
-H "Content-Type: text/plain" \
--data 'log.level="error", log.service="order-service"'
```

#### Expected JSON Output

The service will return a JSON object containing the logs that match your query.

```json
{
  "results": [
    {
      "log.level": "error",
      "log.message": "Failed to process order",
      "log.service": "order-service",
      "log.traceId": "xyz-123"
    }
  ]
}
```

## Development and Testing

To run the service without a live Elasticsearch instance, you can use the mock adapter. This is useful for end-to-end testing of the API and query logic.

Enable the mock adapter using the `--env=test` command-line flag:

```bash
go run ./cmd/mangle-service --env=test
```

When running in test mode, the service will return predefined mock data for any query, allowing you to test the application's behavior in isolation.
