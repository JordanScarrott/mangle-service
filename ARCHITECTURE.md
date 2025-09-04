# Mangle Service Architecture

This document provides a comprehensive overview of the architectural design of the `mangle-service`. It is intended to guide developers in understanding the structure, principles, and patterns employed in this project.

## 1. High-Level Overview

The `mangle-service` is built upon the principles of **Ports and Adapters**, also known as **Hexagonal Architecture**. This architectural pattern is designed to create a clear separation between the application's core business logic and the external services, tools, and delivery mechanisms it interacts with.

The primary motivation for choosing this architecture is to achieve:

- **Decoupling:** The core business logic is completely independent of external concerns like web frameworks, databases, or third-party APIs. This means any external component can be replaced with minimal impact on the core application.
- **Testability:** With the business logic isolated, it can be tested in its entirety without needing to spin up a web server or connect to a real database. We can use simple mocks for our ports, making tests fast, reliable, and easy to write.
- **Maintainability:** The clear boundaries and separation of concerns make the codebase easier to understand, reason about, and extend over time. New features can be added to the core without being tangled up in infrastructure-specific code.

## 2. Core Architectural Concepts

The Hexagonal Architecture can be broken down into three fundamental components:

### The Core Application (The Hexagon)

This is the heart of our application. It contains all the business logic, domain models, and application services. Crucially, the Core has no knowledge of the outside world. It doesn't know about HTTP, JSON, or any specific database. It is a pure, self-contained representation of the application's purpose.

### Ports

Ports are the gateways to the Core Application. In Go, they are defined as `interfaces`. These interfaces form a contract, specifying how external components can interact with the core or how the core can interact with external tools.

Think of them like the USB ports on a computer: they define a standard for connection but have no idea what specific device (a keyboard, a mouse, a flash drive) will be plugged in.

### Adapters

Adapters are the concrete implementations that connect to the Ports. They are the "plugs" that fit into the ports. They are responsible for translating communication between the external world and the core application.

There are two main types of adapters:

-   **Driving Adapters (or Primary Adapters):** These adapters drive the application. They take input from the outside world and call the core's ports. Our HTTP server is a perfect example of a driving adapter.
-   **Driven Adapters (or Secondary Adapters):** These adapters are driven by the application. The core calls out to these adapters through ports to get work done (e.g., fetching data from a database or another API). Our client for the external Mangle service is a driven adapter.

## 3. Visual Diagram

Here is a simplified ASCII diagram representing the architecture of the `mangle-service`:

```
+-----------------------------------------------------------------+
|                        External World                           |
|                                                                 |
|   +------------------+         +----------------------------+   |
|   |   HTTP Client    |         | External Mangle Service    |   |
|   +------------------+         +----------------------------+   |
|           |                              ^                      |
|           v                              |                      |
|  +------------------+           +--------------------------+    |
|  |   HTTP Adapter   |           |   Mangle Client Adapter  |    |
|  |  (Driving)       |           |   (Driven)               |    |
|  +------------------+           +--------------------------+    |
|           |                              ^                      |
+-----------|------------------------------|----------------------+
            |                              |
            v (calls)                      ^ (implements)
      +--------------+             +-----------------+
      | Port         |             | Port            |
      | (Service     |             | (Repository     |
      | Interface)   |             | Interface)      |
      +--------------+             +-----------------+
            |                              ^
            |  +-------------------------+ |
            |  |                           | |
            +->|   Core Application Logic  | |
               |   (The Hexagon)           |<-+
               |                           |
               +---------------------------+
```

## 4. Mapping Architecture to the `mangle-service` Directory Structure

The abstract concepts described above are mapped directly to the project's directory structure:

-   `cmd/server/main.go`: **Application Entry Point**
    This is the "wiring" layer. The `main` function is responsible for instantiating the concrete adapters (like the HTTP server and Mangle client) and injecting them into the core application services. It brings the whole application to life.

-   `/internal/core`: **The Core Application (The Hexagon)**
    This directory is the heart of the service. It contains the pure business logic and domain models, completely isolated from any external technology.
    -   `/internal/core/domain`: Contains the core data structures (structs) that represent the business domain.
    -   `/internal/core/service`: Contains the implementation of the core's business logic, which orchestrates the domain objects.

-   `/internal/core/ports`: **The Ports**
    This directory contains the Go `interface` definitions that form the boundary of our core application. These interfaces are the contracts that adapters must fulfill.

-   `/internal/adapters`: **The Adapters**
    This directory contains the concrete implementations of the ports. They are the bridge between the core and the outside world.
    -   `/internal/adapters/http`: A **Driving Adapter** that handles incoming HTTP requests, validates them, and calls the appropriate service on the core.
    -   `/internal/adapters/mangle`: A **Driven Adapter** that implements the repository port to connect and communicate with the external Mangle service.

## 5. The Dependency Rule

The entire architecture is governed by one strict rule:

> **Dependencies must only point inwards.**

This means:
- The **Adapters** depend on the **Core**.
- The **Core** depends on **nothing** outside of itself (except its own ports).

The Core Application must have zero knowledge of the existence of any adapter. It only knows about the `interfaces` (Ports) it defines. This rule is fundamental to achieving decoupling and is enforced by the Go compiler through the use of interfaces.

## 6. Lifecycle of a Request

To make this concrete, let's trace the journey of a typical API request through the system:

1.  An HTTP request (e.g., `POST /mangle`) hits the server running in `main.go`.
2.  The **HTTP Adapter** (`/internal/adapters/http/server.go`) receives the raw HTTP request.
3.  The adapter's handler function is responsible for parsing and validating the request body (e.g., JSON).
4.  It deserializes the JSON payload into a domain model struct defined in the core (`/internal/core/domain/query.go`).
5.  The adapter then calls a method on a **Port** (e.g., `QueryService.Mangle()`), which is an interface defined in `/internal/core/ports/service.go`.
6.  The **Core Service** (`/internal/core/service/query_service.go`), which implements the `QueryService` interface, receives the call and executes the primary business logic.
7.  During its execution, the core service might need data from an external source. It calls a method on another **Port** (e.g., `MangleRepository.GetMangledString()`), an interface defined in `/internal/core/ports/repository.go`.
8.  The **Mangle Adapter** (`/internal/adapters/mangle/mangle_adapter.go`), which was injected into the core service at startup, implements this repository interface. Its method is invoked.
9.  The Mangle adapter is responsible for making the actual network call to the external Mangle API, handling things like authentication, serialization, and error handling for that specific interaction.
10. The result is returned from the adapter to the core service.
11. The core service completes its logic and returns the final result object to the HTTP adapter.
12. Finally, the **HTTP Adapter** serializes the result object into a JSON response and writes it back to the client with an appropriate HTTP status code.

This clear, unidirectional flow ensures that our core logic remains pure and testable, while the messy details of external communication are neatly contained within the adapters.
