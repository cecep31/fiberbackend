# GEMINI Code Companion: Go Fiber Backend

This document provides a comprehensive guide for developers working on the Go Fiber Backend project. It outlines the project's architecture, development practices, and operational procedures.

## Project Overview

This project is a modern REST API for a blog/content management system built with the Go programming language and the Fiber web framework. It leverages a PostgreSQL database for data storage and follows a clean architecture pattern, separating concerns into distinct layers: handlers, services, and repositories.

### Key Technologies

*   **Backend:** Go (version 1.25 or higher)
*   **Web Framework:** Fiber
*   **Database:** PostgreSQL
*   **Authentication:** JWT (JSON Web Tokens)
*   **Containerization:** Docker
*   **CI/CD:** GitHub Actions

### Architecture

The project adheres to a clean architecture, promoting a separation of concerns and maintainability. The core components are:

*   **`cmd/main.go`**: The application's entry point, responsible for initializing the server, loading configuration, and setting up dependencies.
*   **`internal/`**: Contains the core application logic, organized into the following subdirectories:
    *   **`handler`**: HTTP handlers that process incoming requests and formulate responses.
    *   **`service`**: Business logic and use cases.
    *   **`repository`**: Data access layer that interacts with the PostgreSQL database.
    *   **`model`**: Data structures representing database entities.
    *   **`middleware`**: Custom middleware for handling cross-cutting concerns like authentication and logging.
    *   **`routes`**: Defines the application's API endpoints.
*   **`pkg/`**: Shared utility packages.
*   **`migrations/`**: SQL scripts for database schema management.

## Building and Running

### Prerequisites

*   Go 1.25+
*   PostgreSQL 14+
*   Docker (optional)

### Local Development

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd fiberbackend
    ```

2.  **Set up environment variables:**
    Copy the example environment file and customize it with your local settings:
    ```bash
    cp .env.example .env
    ```

3.  **Install dependencies:**
    ```bash
    go mod download
    ```

4.  **Run the application:**
    For a single run, use:
    ```bash
    go run cmd/main.go
    ```

    For live-reloading during development, the `.air.toml` file is configured to automatically rebuild and restart the application when changes are detected:
    ```bash
    air
    ```

### Testing

Run the test suite with:

```bash
make test
```

## Development Conventions

*   **Code Style:** Adhere to the standard Go formatting guidelines. Use `make fmt` to format the code.
*   **Linting:** The project uses `golangci-lint` for static analysis. Run the linter with `make lint`.
*   **Contribution:** Follow the guidelines outlined in the `README.md` file. All new features should include corresponding tests.

## Deployment

### Containerization

The project includes a multi-stage `Dockerfile` for building a lightweight and secure container image. To build the Docker image:

```bash
docker build -t fiberbackend .
```

To run the application in a Docker container:

```bash
docker run -p 8080:8080 fiberbackend
```

### Continuous Integration and Deployment (CI/CD)

The repository is configured with a GitHub Actions workflow in `.github/workflows/main.yml`. The workflow automates the following process on every push to the `main` branch:

1.  **Build and Push:** A Docker image is built and pushed to Docker Hub.
2.  **Deploy:** The application is deployed to Fly.io.

The deployment process relies on secrets for Docker Hub and Fly.io API tokens, which are configured in the GitHub repository's settings.
