# Bakery API architecture

โค้ดชุดปัจจุบันแยก feature ภายใต้ `internal/usecases` ตามโครงสร้างล่าสุดของโปรเจกต์

```text
internal/
  shared/                     errors, UUID validation
  usecases/
    product/                  catalog service + repository contract
    users/                    customer + verified LINE identity contracts
    orders/                   order/payment aggregate + repository contract
    store/                    store settings contract
    composite/
      checkout.go             validate cross-feature checkout/create
      notifications.go        dispatch persisted outbox via sender port
    createProduct.go, ...     previous bigint product usecases retained
  adapters/
    inbound/api/              versioned HTTP routes and stateless access checks
    inbound/rest/             previous product handlers retained in source
    outbound/db/
      catalog/                UUID product/category adapter
      users/                  users/customers adapter
      orders/                 transactional order/payment/inventory + outbox
      store/                  persistent store settings
      product/                previous bigint product adapter
    outbound/line/            LINE identity verification and message sender
cmd/bakery/wire/               compose dependencies and worker lifecycle
```

Dependency imports flow from adapters/composite to feature contracts and then to shared. Shared imports only the standard library and never depends on product/users/composite. Each feature owns its contract; reuse a feature port across workflows instead of relocating domain-specific repositories into shared.

Product and users are separate folders. Orders/payments/inventory are currently a single transactional aggregate: its repository performs atomic writes using one PostgreSQL transaction. No pgx or HTTP framework types appear in the feature contracts. The DB adapter currently also enforces order-state transitions and reservation rules; extracting that policy into dedicated domain operations is a useful future refactor, not a prerequisite for reviewing this API implementation.

Composite checkout validates input before invoking the aggregate repository. Composite notifications uses separate outbox and sender contracts to dispatch messages after commit. The order approval transaction inserts the outbox row atomically, and the worker lives until the application cleanup callback is called.

Store settings have business ownership and live in `store`, rather than shared. LINE identities are persisted in `users.auth_provider='line'` and `users.provider_subject`. The schema does not contain a separate `users.line_user_id` column.

The new router is wired at `/api/v1` with unversioned aliases and uses the UUID schema. Previous bigint Product handlers remain in source for compatibility review, but are not wired to the server. See [API.md](API.md) for the complete routes, request examples, authorization, migration and review checklist.
