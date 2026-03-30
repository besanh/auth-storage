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
