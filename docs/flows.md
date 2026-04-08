# Auth Storage Service: Sequence Flows

This document details the core authentication and authorization flows between the Client, AuthService, Biz layer, and external storage systems.

## 1. User Registration
Handles user creation in the database and initial provisioning in SpiceDB.

```mermaid
sequenceDiagram
    participant Client
    participant Service
    participant Biz as AuthUseCase
    participant DB as Postgres
    participant SpiceDB

    Client->>Service: Register(email, password)
    Service->>Biz: Register(ctx, email, password)
    Biz->>Biz: bcrypt.HashPassword(password)
    Biz->>DB: Transaction: Insert User
    Biz->>SpiceDB: Provision: Platform Global Relationship
    Biz-->>Service: UserID
    Service-->>Client: RegisterReply(userID)
```

## 2. User Login
Authenticates user credentials and generates Initial JWT pairs.

```mermaid
sequenceDiagram
    participant Client
    participant Service
    participant Biz as AuthUseCase
    participant DB as Postgres

    Client->>Service: Login(email, password)
    Service->>Biz: Login(ctx, email, password)
    Biz->>DB: GetUserByEmail(email)
    DB-->>Biz: User Record
    Biz->>Biz: bcrypt.CompareHash(hash, password)
    Biz->>Biz: GenerateToken(userID)
    Note over Biz: Signs with RSA Key + Configured KID
    Biz-->>Service: AccessToken, RefreshToken
    Service-->>Client: LoginReply(...)
```

## 3. Token Refresh (with Rotation)
Validates old refresh token, checks blacklist, and provides a completely new pair.

```mermaid
sequenceDiagram
    participant Client
    participant Service
    participant Biz as AuthUseCase
    participant Redis
    participant DB as Postgres

    Client->>Service: RefreshToken(old_refresh_token)
    Service->>Biz: RefreshToken(ctx, old_refresh_token)
    Note over Biz: Step 1: Validate Refresh Token
    Biz->>Biz: jwt.ParseWithClaims(old_token)
    Note over Biz: Fixed: Passing pointer to claims
    Note over Biz: Step 2: Check Blacklist
    Biz->>Redis: IsTokenBlacklisted(JTI)
    Redis-->>Biz: (bool)
    alt is blacklisted
        Biz-->>Service: error: Invalid Token
        Service-->>Client: Error reply
    else is valid
        Note over Biz: Step 3: Extract User ID
        Note over Biz: Step 4: Security Check
        Biz->>DB: GetUserByID(userID)
        DB-->>Biz: User Record
        Note over Biz: Step 5: Token Rotation
        Biz->>Biz: GenerateToken(userID)
        Biz-->>Service: NEW AccessToken, NEW RefreshToken
        Service-->>Client: RefreshTokenReply(...)
    end
```

## 4. User Logout (Blacklisting)
Invalidates a refresh token by adding its ID to Redis.

```mermaid
sequenceDiagram
    participant Client
    participant Service
    participant Biz as AuthUseCase
    participant Redis

    Client->>Service: Logout(refresh_token)
    Service->>Biz: Logout(ctx, refresh_token)
    Biz->>Biz: parser.ParseUnverified(token)
    Biz->>Biz: Get JTI & Remaining TTL
    Biz->>Redis: SetEX(blacklist:JTI, TTL, JTI)
    Biz-->>Service: success
    Service-->>Client: LogoutReply(message)
```

## 5. Machine-to-Machine (M2M) Login
Authenticates microservices or automated clients using static IDs and secrets.

```mermaid
sequenceDiagram
    participant Client as External Service
    participant Service as M2MAuthService
    participant Biz as M2MAuthUseCase
    participant DB as Postgres

    Client->>Service: Login(client_id, client_secret)
    Service->>Biz: Login(ctx, client_id, client_secret)
    Biz->>DB: GetMachineClientByID(client_id)
    DB-->>Biz: Client Record (includes hashed secret)
    Biz->>Biz: bcrypt.CompareHash(hash, client_secret)
    
    rect rgb(240, 240, 240)
    Note over Biz: Generate M2M Token
    Biz->>Biz: Set claims: sub=client_id, type=m2m
    Biz->>Biz: Sign with RSA Private Key
    end

    Biz-->>Service: M2M AccessToken, ExpiresIn
    Service-->>Client: LoginReply(...)
```

## 6. Business Operations & API Ordering

This section describes the order in which internal and external APIs must be called to ensure data and permission consistency.

### 6.1 Creating a Standard Folder
When a user creates a new top-level folder.

1.  **App Logic**: Generate a new UUID for the folder.
2.  **App Logic**: Insert the folder record into the SQL database.
3.  **Auth Call**: `Permission.WriteRelationship`
    *   `resource_type`: `"folder"`
    *   `resource_id`: `[folder_uuid]`
    *   `relation`: `"owner"`
    *   `subject_type`: `"user"`
    *   `subject_id`: `[user_id]`

### 6.2 Creating a Folder in a Shared Directory
When a user creates a folder inside another folder that has its own permissions.

1.  **App Logic**: Generate a new UUID for the folder.
2.  **App Logic**: Insert the folder record into the SQL database (linked to parent).
3.  **Auth Call (Owner)**: `Permission.WriteRelationship`
    *   `relation`: `"owner"`, `subject_type`: `"user"`, `subject_id`: `[user_id]`
4.  **Auth Call (Inheritance)**: `Permission.WriteRelationship`
    *   `resource_type`: `"folder"`
    *   `resource_id`: `[new_folder_uuid]`
    *   `relation`: `"parent"`
    *   `subject_type`: `"folder"`
    *   `subject_id`: `[parent_folder_uuid]`
    *   *Note: This ensures parent permissions flow down to the new child.*

### 6.3 Verifying Access before Action
Before allowing a user to read or modify a file/folder.

1.  **Auth Call**: `Permission.CheckPermission`
    *   `resource_type`: `"folder"`
    *   `resource_id`: `[target_id]`
    *   `relation`: `"read"` (or `"write"` / `"delete"`)
    *   `subject_type`: `"user"`
    *   `subject_id`: `[user_id]`
2.  **App Logic**: Proceed with the database operation ONLY if `Allowed: true`.


