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
  - service: "api-gateway"
    depends_on:
      - "order-service"
```
This configuration generates a `calls("api-gateway", "order-service")` fact within Mangle.

### Step 2: Write a Mangle Query

Queries are sent as a JSON object in an API request. Let's write a query to find all services that have logged a 500 error.

```mangle
logs(_, Service, 500, _).
```

### Step 3: Execute the Query

Use `curl` to send the query to the running service.

```bash
curl -X POST http://localhost:8080/query \
-H "Content-Type: application/json" \
--data '{"query": "logs(_, Service, 500, _)."}'
```

#### Expected JSON Output

The service will return a JSON object containing the bindings for the variables in your query.

```json
{
    "count": 1,
    "results": [
        {
            "Service": "order-service"
        }
    ]
}
```

## Advanced Usage: Debugging a Cascading Failure

This new section should be placed after the 'Quick Start Guide' and before 'Development and Testing'. It must walk the user through a realistic and powerful debugging scenario.

### The Scenario

Imagine your `api-gateway` is failing with 500 errors. Your initial instinct is to check the gateway's logs, but the error message is generic, like 'Internal Server Error'. The real problem might be hidden in a downstream service. This is where Mangle excels.

### The Goal

Our goal is to write a query that finds the original service that errored within the same transaction that caused the gateway to fail.

### Step-by-Step Guide

#### Step 1: Model Your Service Relationships

Reiterate the importance of the `relationships.yml` file. Show the simple configuration needed for this scenario:

```yaml
relationships:
  - service: "api-gateway"
    depends_on:
      - "order-service"
```

#### Step 2: Write the Investigative Query

Present the Mangle query. **Crucially, explain each part of the query in plain English** so the user understands the logic. The following logic is sent as a single query string.

*Part 1: First, we create a temporary rule to find all Trace IDs where the 'api-gateway' service logged a 500 error.*
```mangle
gateway_crashed(TraceID) :- logs(TraceID, "api-gateway", 500, _).
```

*Part 2: Now, we find the root cause. This rule reads: "A service is the root cause IF... it belongs to a transaction where the gateway crashed (using our rule above), AND the gateway is known to call that service, AND that service ALSO logged a 500 error on the exact same trace."*
```mangle
root_cause_service(Service, TraceID) :- gateway_crashed(TraceID), calls("api-gateway", Service), logs(TraceID, Service, 500, _).
```

#### Step 3: Execute and Find the Truth

Show the `curl` command to run the `root_cause_service(Service, TraceID)` query. Note that each rule must end with a period.

```bash
curl -X POST http://localhost:8080/query \
-H "Content-Type: application/json" \
--data '{"query": "gateway_crashed(TraceID) :- logs(TraceID, \"api-gateway\", 500, _). root_cause_service(Service, TraceID) :- gateway_crashed(TraceID), calls(\"api-gateway\", Service), logs(TraceID, Service, 500, _). root_cause_service(Service, TraceID)."}'
```

Display the expected JSON output, which should clearly point to the `order-service`.
```json
[
  {
    "Service": "order-service",
    "TraceID": "trace-xyz"
  }
]
```

#### Step 4: The Insight

Conclude with a powerful summary: "This result immediately directs the on-call engineer to investigate the `order-service`, not the `api-gateway`, saving critical time and preventing misdiagnosis of the issue."

## Development and Testing

To run the service without a live Elasticsearch instance, you can use the mock adapter. This is useful for end-to-end testing of the API and query logic.

Enable the mock adapter using the `--env=test` command-line flag:

```bash
go run ./cmd/mangle-service --env=test
```

When running in test mode, the service will return predefined mock data for any query, allowing you to test the application's behavior in isolation.
