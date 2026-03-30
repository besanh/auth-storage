# Auth Storage Service: Architecture Overview

The Auth Storage service is a Go-based authentication and authorization provider built using the **Kratos** framework. It provides secure identity management, token-based authentication (JWT), and fine-grained authorization integration.

## Technology Stack

- **Framework**: [Kratos](https://go-kratos.dev/) for gRPC and HTTP trans-coding.
- **Database**: **Postgres** for persistent user storage, managed via **SQLC**.
- **Authorization**: **SpiceDB** (Authzed) for relationship-based access control (ReBAC).
- **Caching/Blacklisting**: **Redis** for managing blacklisted JWT tokens during logout.
- **DI**: **Google Wire** for compile-time dependency injection.

## Core Components

- **API Layer (`/api/auth/v1`)**: Protocol buffers defining the `Auth` service contracts and REST mappings.
- **Business Layer (`internal/biz`)**: Domain logic including JWT generation, password hashing, and token rotation.
- **Data Layer (`internal/data`)**: Repository implementations for Postgres, Redis client initialization, and SpiceDB client integration.
- **Service Layer (`internal/service`)**: Implements gRPC handlers and maps incoming requests to usecase calls.

## Key Features

### 1. JWT with Dynamic KID
Authenticates users using RSA-signed JWTs. The Key ID (**KID**) is dynamically generated via the **Makefile** and loaded by the application at startup. This enables safe key rotation by naming private/public keys with their unique ID.

> [!CAUTION]
> **Security Warning**: Private keys (`*.pem`) should **never** be committed to version control. They are ignored via `.gitignore`. In production, these should be managed through secure secrets storage (e.g., Kubernetes Secrets, AWS Secrets Manager).

### 2. Token Rotation
Every refresh token request results in a **brand new** Access Token and Refresh Token pair. This "rotation" strategy improves security by reducing the window of opportunity for stolen tokens.

### 3. SpiceDB Integration
During registration, users are automatically provisioned into **SpiceDB**. High-level "platform global" relationships are established, allowing for future fine-grained permission checks.

### 4. Redis-backed Logout
When a user logs out, their Refresh Token's unique ID (**JTI**) is stored in **Redis** as a blacklisted entry. The entry's TTL is set to the token's remaining expiration time, ensuring the token cannot be reused even before it naturally expires.

### 5. Visual API Documentation (Swagger UI)
The project includes a **Swagger UI** integration via Docker Compose. Developers can interact with the API endpoints visually by running `make swagger`, which serves the OpenAPI specification at `http://localhost:8080`.
