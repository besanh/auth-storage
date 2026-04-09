# Integration Guide: Using Auth Storage Service

This guide explains how external microservices should integrate with the Auth Storage service for authentication and authorization.

## 1. Authentication

### User Authentication
To authenticate an end-user, redirect them or call the Login API:
*   **Endpoint**: `POST /api/auth/v1/login`
*   **Returns**: `access_token`, `refresh_token`.
*   **Usage**: Include the `access_token` in the `Authorization: Bearer <token>` header for all Subsequent requests.

### Machine-to-Machine (M2M) Authentication
Used by services (like `file-service`) to authenticate with each other.
*   **Endpoint**: `POST /api/m2m_auth/v1/login`
*   **Payload**: `{"client_id": "...", "client_secret": "..."}`
*   **Usage**: The returned token has a `type: m2m` claim and is used for internal service-to-service calls.

---

## 2. Authorization (SpiceDB)

Other services use the `Permission` API to check if a user is allowed to perform an action.

### Checking Permissions
Before performing a sensitive operation (e.g., deleting a file):
*   **Endpoint**: `POST /v1/internal/permissions/check`
*   **Payload**:
```json
{
    "resource_type": "folder",
    "resource_id": "folder-uuid",
    "relation": "delete",
    "subject_type": "user",
    "subject_id": "user-uuid"
}
```
*   **Result**: `{"allowed": true/false}`

### Provisioning Resources
When a service creates a new resource (e.g., a Folder), it must register the ownership in SpiceDB:
*   **Endpoint**: `POST /v1/internal/permissions/write`
*   **Payload** (Make user the owner):
```json
{
    "resource_type": "folder",
    "resource_id": "new-folder-uuid",
    "relation": "owner",
    "subject_type": "user",
    "subject_id": "creator-user-uuid"
}
```

---

## 3. Best Practices

1.  **Extracting Subject**: In your Go code, use `biz.ExtractSubject(ctx)` to safely retrieve the `sub` (UserID or ClientID) from the incoming JWT without worrying about the underlying claim type.
2.  **Inheritance**: When creating resources inside a container (e.g., a file inside a folder), write a `parent` relationship to SpiceDB so permissions automatically flow down.
3.  **Token Rotation**: Always handle `401 Unauthorized` by attempting to use the `refresh_token` before forcing a user logout.

---

## 4. Link-Based Authorization (Shared Links)

The `ShareService` allows creating public or restricted links for resources. These are validated using the `Permission` API by treating the link token as the `subject`.

### Using a Share Link
When a request arrives with a `token` (e.g., from a query parameter or a specific header):

1.  **Extract the Token**: Identify the link token provided by the user.
2.  **Verify via Permission Service**:
    *   **Endpoint**: `POST /v1/internal/permissions/check`
    *   **Payload**:
    ```json
    {
        "resource_type": "file",
        "resource_id": "file-uuid",
        "relation": "viewer",
        "subject_type": "share_link",
        "subject_id": "LINK_TOKEN_HERE"
    }
    ```
3.  **Handle Success**: If `allowed: true`, proceed with the requested operation.

### Managing Links
*   **Create**: Use `POST /v1/shares/create` to generate a new token and automatically register it in SpiceDB.
*   **Revoke**: Use `POST /v1/shares/revoke` to delete the token. Access will be immediately denied by the Permission service.
*   **Update**: Use `POST /v1/shares/update` to change the permission level (e.g., from `viewer` to `editor`) for an existing link.
