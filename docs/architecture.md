# Auth Storage Service: Architecture Overview

The Auth Storage service is a Go-based authentication and authorization provider built using the **Kratos** framework. It provides secure identity management, token-based authentication (JWT), and fine-grained authorization integration.

## Technology Stack

- **Framework**: [Kratos](https://go-kratos.dev/) for gRPC and HTTP trans-coding.
- **Database**: **Postgres** for persistent user storage, managed via **SQLC**.
- **Authorization**: **SpiceDB** (Authzed) for relationship-based access control (ReBAC).
- **Caching/Blacklisting**: **Redis** for managing blacklisted JWT tokens during logout.
- **DI**: **Google Wire** for compile-time dependency injection.

## Core Components

- **API Layer (`/api`)**: Protocol buffers defining service contracts.
    - `auth/v1`: Generic authentication and token management.
    - `m2m_auth/v1`: Machine-to-machine authentication.
    - `permission/v1`: Internal API for permission checks and relationship management.
- **Service Layer (`internal/service`)**: Implements gRPC/HTTP handlers and maps proto requests to domain-specific usecase calls.
- **Business Layer (`internal/biz`)**: Protocol-agnostic domain logic and repository interfaces. Houses core logic like JWT generation, password hashing, and authorization orchestration.
- **Data Layer (`internal/data`)**: Infrastructure implementations, including database queries, SpiceDB client integration, and Redis operations.

## Key Features

### 1. Protocol-Agnostic Core
The `biz` and `data` layers are completely decoupled from protocol buffers. This ensures that the core business logic is independent of whether it's accessed via gRPC, HTTP, or internal calls.

### 2. Centralized Transaction Management
Cross-repository atomic operations are managed via a `Transaction` manager in the `data` layer. It utilizes `context` to propagate `*sql.Tx` across data access methods, ensuring data consistency between the relational database and external systems like SpiceDB.

### 3. SpiceDB Integration
Fine-grained authorization is handled by SpiceDB. The `Permission` service provides a simplified internal interface for other microservices to check permissions or update ACL relationships.

### 4. JWT with Dynamic KID
Authenticates users using RSA-signed JWTs. The Key ID (**KID**) is dynamically loaded at startup, enabling safe key rotation.

### 5. Redis-backed Logout
Managed blacklisting of JWT tokens ensures that logout is immediate and global across the system.
