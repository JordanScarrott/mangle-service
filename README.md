Mangle Query Service
Mangle Query Service is a backend Go application that provides a simplified HTTP API for querying log data using Google's Mangle query language. It is built using a clean, maintainable Ports and Adapters (Hexagonal) Architecture.
This architecture isolates the core application logic from external dependencies and implementation details, making the service easier to test, maintain, and evolve.
Architecture Overview
The project follows a strict Ports and Adapters pattern to ensure a clean separation of concerns.
 * Core Application (/internal/core): This is the heart of the service. It contains the business logic, domain models, and service interfaces (ports). It has zero knowledge of external technologies like databases, APIs, or web frameworks.
 * Ports (/internal/core/ports): These are Go interfaces that define the contracts for interacting with the core logic (driving ports) and for the core logic to interact with external tools (driven ports).
 * Adapters (/internal/adapters): These are the concrete, interchangeable implementations of the ports.
   * HTTP Adapter: A primary/driving adapter that exposes the application's functionality via a RESTful API. It translates incoming HTTP requests into calls to the core service.
   * Mangle Adapter: A secondary/driven adapter that implements the repository port. It handles all the specific details of communicating with the Google Mangle API to execute queries.
Directory Structure
/mangle-service
├── cmd
│   └── server
│       └── main.go         # Application entry point and dependency injection.
├── internal
│   ├── core
│   │   ├── domain          # Core business models (e.g., Query, QueryResult).
│   │   ├── ports           # Service and repository interfaces.
│   │   └── service         # Core business logic implementation.
│   └── adapters
│       ├── http            # HTTP server, request handlers, and routing.
│       └── mangle          # Google Mangle client adapter.
├── pkg
│   └── logger              # Shared logging utility.
└── go.mod

Getting Started
Follow these instructions to get the service running on your local machine for development and testing.
Prerequisites
 * Go: Version 1.21 or later.
 * Google Cloud SDK: The gcloud CLI, configured and authenticated.
 * Application Default Credentials (ADC): You must be authenticated with Google Cloud. Run the following command to set up your credentials:
   gcloud auth application-default login

Installation & Running Locally
 * Clone the repository:
   git clone https://github.com/JordanScarrott/mangle-service.git
cd mangle-service

 * Install dependencies:
   The Go module system will handle dependencies automatically. To install them explicitly, you can run:
   go mod tidy

 * Run the service:
   The application can be started from the cmd/server directory.
   go run ./cmd/server/main.go

   By default, the server will start on port 8080.
Running Tests
To run all unit and integration tests, execute the following command from the root directory:
go test ./...

API Usage
The service exposes a simple endpoint for executing Mangle queries.
Execute a Query
 * Endpoint: POST /query
 * Description: Submits a Mangle query for execution.
 * Request Body:
   {
  "query": "s.log_id='your_log_id' | limit 10"
}

 * Example curl Request:
   curl -X POST http://localhost:8080/query \
-H "Content-Type: application/json" \
-d '{
  "query": "s.log_id='\''cloudaudit.googleapis.com%2Factivity'\'' and s.timestamp > now() - 1h | limit 5"
}'

 * Success Response (200 OK):
   {
  "results": [
    {
      "field1": "value1",
      "field2": "value2"
    }
  ],
  "count": 1
}

 * Error Response (400 Bad Request or 500 Internal Server Error):
   {
  "error": "A descriptive error message."
}
