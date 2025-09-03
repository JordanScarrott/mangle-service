# Mangle Go Microservice Example

## Purpose

This project demonstrates how to use the [Google Mangle](https://github.com/google/mangle) library in a simple Go microservice. It shows how to:

*   Structure a Go application with `cmd` and `internal` directories.
*   Define Datalog rules and facts.
*   Load rules and facts into a Mangle engine.
*   Run queries against the data.
*   Containerize the application using Docker.

## How to Run

### Prerequisites

*   Go (version 1.18 or later)
*   Docker (optional, for containerized version)

### Running Locally

1.  **Clone the repository (or have the files on your local machine).**

2.  **Build and run the application:**

    ```sh
    go run ./cmd/mangle-microservice
    ```

### Running with Docker

1.  **Build the Docker image:**

    ```sh
    docker build -t mangle-microservice .
    ```

2.  **Run the Docker container:**

    ```sh
    docker run --rm mangle-microservice
    ```

## Expected Output

When you run the application, you should see the following output:

```
Query: depends_on(X, "service-c")?
Results:
  X = "service-b"
  X = "service-a"
```

This output shows that `service-a` and `service-b` both have a dependency on `service-c`, as defined by the `depends_on` rule. The order of the results may vary.
