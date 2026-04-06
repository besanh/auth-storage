# Developer Guide: Transactions and Layering

This guide explains how to properly use the transaction management system and maintain the protocol-agnostic nature of the `biz` and `data` layers.

## 1. Using Transactions

The `biz.Transaction` interface allows usecases to orchestrate atomic operations across multiple repositories.

### Example: Usecase Implementation

```go
func (uc *MyUseCase) DoSomethingAtomic(ctx context.Context, data *MyData) error {
    return uc.tm.ExecTx(ctx, func(ctx context.Context) error {
        // Method 1: Repo update
        if err := uc.repo.Save(ctx, data); err != nil {
            return err
        }
        
        // Method 2: Another repo update (using same transaction via context)
        if err := uc.otherRepo.Link(ctx, data.ID); err != nil {
            return err
        }
        
        return nil
    })
}
```

### Data Layer Implementation (SQLC Example)

Repositories should use the `FromContext` helper to check for an existing transaction:

```go
func (r *myRepo) Save(ctx context.Context, d *MyData) error {
    q := r.data.Query
    if tx, ok := FromContext(ctx); ok {
        q = q.WithTx(tx)
    }
    return q.Insert(ctx, ...)
}
```

## 2. Protocol Decoupling

Core logic and data storage must **never** import or use generated proto types (`pb`).

### Pattern
1.  **Service Layer**: Receives proto request, maps to domain struct, calls Usecase.
2.  **Biz Layer**: Performs logic using domain structs, calls Repositories.
3.  **Data Layer**: Receives domain structs, maps them to infrastructure-specific types (e.g., SQL params, SpiceDB protos), and executes.

### Why?
-   **Stability**: Domain logic doesn't break when API contracts change.
-   **Testability**: Usecases can be tested without mocking complex proto stubs.
-   **Reusability**: `biz` logic can be used by internal jobs or CLI tools without overhead.
