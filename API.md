# TongBai REST API — review contract

Base path: `/api/v1`. UUID identifiers match `app_schema`. Database money uses integer satang; the frontend DTO uses `*_thb` (converted at the adapter boundary). There is no server-side session store.

## Routes

| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/health` | public | Process health (not DB readiness) |
| GET | `/store` | public | Store description/delivery settings |
| GET | `/categories` | public | Active categories |
| GET | `/products` | public | Active catalog; `category`, `search`, `limit`, `offset` |
| GET | `/products/{slug-or-id}` | public | Product with image URLs |
| POST | `/checkout/quote` | public | Authoritative price and available-stock check |
| POST | `/orders` | guest/user | Atomic customer/order/items/payment/reservations |
| GET | `/orders` | order/user token | One guest order, or signed-in user's order history |
| GET | `/orders/{id}` | owner/admin | Order details |
| POST | `/payments/{id}/slip` | owner/admin | Multipart field `slip`, JPEG/PNG/WebP, max 5 MB |
| POST | `/payments/{id}/submit` | owner/admin | Submit payment with uploaded evidence |
| GET | `/admin/dashboard` | admin | Sales/order/pending-payment totals |
| GET | `/admin/products` | admin | Catalog including inactive products |
| POST | `/admin/products` | admin | Create product |
| PATCH | `/admin/products/{id}` | admin | Partial product update |
| DELETE | `/admin/products/{id}` | admin | Soft delete with admin actor |
| GET | `/admin/orders` | admin | Paginated orders, optional `status` |
| PATCH | `/admin/orders/{id}` | admin | Update `internal_note` |
| PATCH | `/admin/orders/{id}/status` | admin | Validated fulfillment transition |
| GET | `/admin/payments/{id}/slip` | admin | Private evidence bytes for manual review |
| POST | `/admin/payments/{id}/approve` | admin | Manual payment approval, stock commit and LINE outbox |
| POST | `/admin/payments/{id}/reject` | admin | Reject with required `reason` |
| GET | `/admin/customers` | admin | Customer name/phone/order counts |
| GET | `/admin/settings` | admin | Store settings |
| PATCH | `/admin/settings` | admin | Partial store settings update |
| POST | `/users/line/link` | verified LINE token | Register/link LINE identity and issue user token |
| GET | `/users/me` | user token | Current user ID |
| POST | `/webhooks/line` | LINE signature | Verify and acknowledge webhook; no bot actions yet |

All list endpoints default to `limit=20`, maximum 100, `offset=0`. Public categories return all active categories.

## Authentication and checkout

Set `AUTH_SECRET` to at least 32 random characters. `POST /orders` requires a unique `Idempotency-Key` of 16–100 characters (use a UUID). Retrying the same key and payload returns the same order. Reusing a key with a different payload or user returns 409.

```json
{
  "customer": {"name": "Test Buyer", "phone": "0812345678"},
  "fulfillment": {"type": "delivery", "address": "Bangkok"},
  "items": [{"productId": "PRODUCT_UUID", "quantity": 2}],
  "paymentMethod": "bank_transfer"
}
```

For quote, send the same `items` and `fulfillment` fields. For pickup use `type=pickup` and an ISO timestamp in `pickupAt`; delivery fee is zero. Duplicate product IDs in a request are rejected; aggregate quantities in the client first.

Create returns `id`, `order_number`, `payment_id`, totals, status, items, and an `access_token` scoped to this order, valid for seven days. Send `Authorization: Bearer <access_token>` for subsequent order/payment operations. The signed token permits stateless access, but the client still needs to retain the token or log in again via LINE. Signing-key changes invalidate existing tokens.

With LINE login, call `/users/line/link` with `{"access_token":"LINE_LOGIN_ACCESS_TOKEN"}`. The server verifies channel ID and expiry and reads the profile directly from LINE. Its returned user token is valid for one hour. Send it when creating orders to associate them with that user; `/orders` then returns that user's history. Existing guest orders are not claimed automatically.

For admin calls set `ADMIN_API_TOKEN` and `ADMIN_USER_ID` (an existing active UUID with role `admin`) in the server environment. Send `Authorization: Bearer <ADMIN_API_TOKEN>` from a trusted admin client. This is a bootstrap credential, not a complete staff login system. Never publish it in a `NEXT_PUBLIC_*` variable or public bundle. Startup validates the configured admin; rotate the token/restart after changing admin access.

## Product writes

```json
{
  "category_id": "CATEGORY_UUID",
  "name": "Butter Croissant",
  "slug": "butter-croissant",
  "description_th": "ครัวซองต์เนยสด",
  "price_thb": 95,
  "compare_at_price_thb": null,
  "stock_quantity": 12,
  "is_active": true,
  "is_featured": true
}
```

PATCH accepts these fields individually, including explicit `false`, zero, and nullable comparison price. A stock update below active reservations returns 409. Nonempty `images` in product writes are currently rejected: catalog reads existing `product_images` metadata and builds URLs using `PRODUCT_IMAGE_BASE_URL`. Binary product-image storage/upload needs a storage-provider implementation. The current schema has no `sku` column; do not send `sku`.

## Payment and LINE behavior

Reservations expire after 30 minutes. Quote checks remaining availability; create locks products in deterministic ID order to prevent overselling. Submit requires evidence and a live reservation. Approval requires all reservations to remain live, deducts stock exactly once, commits reservation rows, and records history in one transaction. Expired reservations are ignored for availability, but this version does not automatically mark their order/payment rows expired.

Approval is a manual admin attestation that the paid amount equals the expected amount. Uploading a slip does not verify a bank payment. There is no automatic bank/PromptPay payment provider in this change.

Cancellation is allowed before approval and releases reservations. Once approved, the supported transitions are `confirmed → preparing → ready_for_pickup/shipped → completed`; refund/cancellation of an approved order needs a separate refund workflow. PATCH status body: `{"status":"preparing","note":"Preparing now"}`.

After approval, an outbox record is inserted in the same transaction. If `LINE_CHANNEL_ACCESS_TOKEN` and `LINE_ADMIN_USER_ID` are configured, a worker pushes a store-owner notification every 30 seconds and retries failed messages with backoff. `X-Line-Retry-Key` uses the outbox UUID. Duplicate approvals do not create duplicate outbox rows. The customer-recipient notification flow is not enabled; this version notifies the configured store owner/group.

`LINE_LOGIN_CHANNEL_ID` is for verified customer identity. `LINE_CHANNEL_SECRET` verifies `X-Line-Signature` on webhooks. The receiver currently verifies/acknowledges events without applying bot actions. LINE secrets stay server-side. Login and Messaging API channels should use the same LINE provider if you later send notifications to customer identities.

## Setup and review

Migration `db/migrations/20261010120000_api_settings_and_evidence.sql` adds persistent store settings, private slip bytes, the request hash, and LINE outbox. It has been applied to local Docker PostgreSQL with user approval; dbmate refreshed `db/schema.sql`. Supabase has not been migrated. For another environment, review/apply it with dbmate and configure `.env.example` before starting the API.

The previous bigint Product handlers remain in source with their tests but are no longer wired: the new catalog reads `app_schema.products`. The unversioned aliases now use the new contract as well; old `PUT /products/{id}` clients must migrate to admin PATCH.

Frontend integration is implemented in `../Bakery-frontend/web`: checkout sends a stable `Idempotency-Key`, saves order tokens in tab-scoped sessionStorage, and attaches bearer credentials for payment/history. Admin enters the bootstrap token manually; it stays in memory, not a public env variable. Mock mode is explicit opt-in and never hides live API failures. See that frontend's README for setup and remaining LINE OAuth/payment-instruction work.

Error format: `{"error":{"code":"conflict","message":"conflict"}}`. Invalid input 400; unauthorized access 403; missing record 404; invalid transition/stock/idempotency conflict 409; internal failures 500 without database/token details.

Checks:

```powershell
go test ./...
go vet ./...
# Disposable fresh PostgreSQL DB ending in _test only:
$env:TEST_DATABASE_URL='postgresql://.../bakery_api_review_test?sslmode=disable'
go test ./internal/adapters/inbound/api -run TestCheckoutPostgres -v -count=1
```

Integration tests load the schema snapshot, apply the API migration only if its tables are absent, and seed test-only store settings. They exercise authentication, checkout replay/conflict, concurrent stock reservation, slips, approval deduplication, LINE outbox insertion, cancellation and transaction rollback. They refuse non-test databases and fail if schema already exists, rather than dropping tables.
