# Olive & Oven — Database Design

ฐานข้อมูลสำหรับระบบร้านเบเกอรี รองรับสินค้า ลูกค้า ออเดอร์ การชำระเงิน สต็อก และการตรวจสอบย้อนหลัง

## 1. Technology และหลักการ

- **PostgreSQL** เป็นฐานข้อมูลหลัก
- **S3 / Cloudflare R2** ใช้เก็บรูปสินค้าและหลักฐานการชำระเงิน
- **Redis (optional)** ใช้สำหรับ cache, rate limit และ queue
- Backend API เป็น source of truth สำหรับราคา สต็อก สิทธิ์ และสถานะออเดอร์
- ไม่ใช้ `localStorage` เป็นฐานข้อมูลจริง และไม่เก็บไฟล์รูปภาพเป็น base64 ใน PostgreSQL
- เก็บเวลาเป็น UTC และใช้ `timestamp with time zone` (`timestamptz`)
- จำนวนเงินใช้ `bigint` หน่วยสตางค์ เช่น `฿100.00 = 10000`

## 2. Schema

ต้องสร้าง schema ก่อน migration อื่น:

```sql
CREATE SCHEMA IF NOT EXISTS app_schema;
```

ตารางทั้งหมดอยู่ภายใต้ `app_schema` เช่น `app_schema.products` และ `app_schema.orders`

## 3. Audit columns มาตรฐาน

ตารางข้อมูลหลักให้มีคอลัมน์ชุดนี้:

```sql
created_at timestamp with time zone NOT NULL DEFAULT now(),
created_by uuid,
updated_at timestamp with time zone NOT NULL DEFAULT now(),
updated_by uuid,
deleted_at timestamp with time zone,
deleted_by uuid
```

| Column | รายละเอียด |
|---|---|
| `created_at` | วันที่และเวลาที่สร้างข้อมูล |
| `created_by` | ผู้สร้างข้อมูล อาจเป็น `NULL` ถ้าระบบสร้างเอง |
| `updated_at` | วันที่และเวลาที่แก้ไขล่าสุด |
| `updated_by` | ผู้แก้ไขล่าสุด อาจเป็น `NULL` ถ้าระบบแก้เอง |
| `deleted_at` | วันที่และเวลาที่ทำ soft delete |
| `deleted_by` | ผู้ที่ทำ soft delete |

กฎการใช้งาน:

- `created_by`, `updated_by`, `deleted_by` อ้างอิง `users.id` แบบ nullable
- เมื่อ `deleted_at` มีค่า ต้องมี `deleted_by` ด้วย
- รายการที่ถูกลบให้ใช้ soft delete และไม่แสดงใน query ปกติ
- ไม่ควรลบข้อมูลธุรกรรมด้วย `DELETE`
- `audit_logs`, `order_status_history` และ `inventory_movements` เป็น append-only จึงไม่ต้องมี soft delete
- ให้ backend หรือ database trigger อัปเดต `updated_at` ทุกครั้งที่มีการแก้ไข

```sql
CHECK (
  (deleted_at IS NULL AND deleted_by IS NULL)
  OR
  (deleted_at IS NOT NULL AND deleted_by IS NOT NULL)
)
```

## 4. ความสัมพันธ์หลัก

```text
users ── 0..1 ── customers ──< orders ──< order_items >── products
                         │          │
                         │          ├── payments ──< payment_evidence
                         │          ├── order_status_history
                         │          └── inventory_reservations
categories ──< products ──< product_images
products ──< inventory_movements
orders ──< audit_logs
```

## 5. ตารางผู้ใช้และลูกค้า

### `users`

| Column | Type | รายละเอียด |
|---|---|---|
| `id` | uuid PK | รหัสผู้ใช้ |
| `email` | varchar(255) | unique, nullable |
| `role` | varchar(20) | `customer`, `staff`, `admin` |
| `auth_provider` | varchar(30) | `line`, `email`, `google` |
| `provider_subject` | varchar(255) | รหัสจาก provider |
| `is_active` | boolean | default `true` |
| audit columns | — | ตามมาตรฐานด้านบน |

ห้ามเก็บ access token หรือ refresh token ในตารางนี้

### `customers`

| Column | Type | รายละเอียด |
|---|---|---|
| `id` | uuid PK | รหัสลูกค้า |
| `user_id` | uuid FK nullable | unique; รองรับ guest customer |
| `display_name` | varchar(160) | ชื่อลูกค้า |
| `phone` | varchar(40) | เบอร์โทร; ควรเข้ารหัสตาม policy |
| `line_display_name` | varchar(160) | nullable |
| audit columns | — | ตามมาตรฐานด้านบน |


## 6. ตารางสินค้า

### `categories`

`id uuid PK`, `slug varchar(160) UNIQUE`, `name_th varchar(160)`, `sort_order integer`, `is_active boolean`, audit columns

### `products`

| Column | Type | รายละเอียด |
|---|---|---|
| `id` | uuid PK | |
| `category_id` | uuid FK | หมวดหมู่สินค้า |
| `name` | varchar(200) | ชื่อสินค้า |
| `slug` | varchar(220) UNIQUE | ใช้สร้าง URL เช่น `chocolate-chip-cookie` |
| `description_th` | text | รายละเอียดสินค้า |
| `price_satang` | bigint | ราคาปัจจุบัน หน่วยสตางค์ |
| `compare_at_price_satang` | bigint nullable | ราคาเดิม |
| `stock_quantity` | integer | ยอดสต็อกปัจจุบันแบบ cached |
| `is_active` | boolean | ซ่อนสินค้าเมื่อเป็น `false` |
| `is_featured` | boolean | สินค้าแนะนำ |
| audit columns | — | ตามมาตรฐานด้านบน |

```sql
CHECK (price_satang >= 0),
CHECK (compare_at_price_satang IS NULL OR compare_at_price_satang >= price_satang),
CHECK (stock_quantity >= 0)
```

สินค้าที่เคยขายแล้วให้ทำ `deleted_at` หรือ `is_active = false` แทนการลบจริง

### `product_images`

`id`, `product_id FK`, `object_key`, `mime_type`, `alt_text`, `sort_order`, `is_primary`, audit columns

```sql
CREATE UNIQUE INDEX ux_product_one_primary_image
ON app_schema.product_images(product_id)
WHERE is_primary = true AND deleted_at IS NULL;
```

## 7. ตารางออเดอร์

### `orders`

| Column | Type | รายละเอียด |
|---|---|---|
| `id` | uuid PK | internal ID |
| `order_number` | varchar(40) UNIQUE | เลขออเดอร์ เช่น `OO-260910-184` |
| `idempotency_key` | varchar(100) UNIQUE | ป้องกันสร้างออเดอร์ซ้ำ |
| `customer_id` | uuid FK nullable | รองรับ guest order |
| `customer_name_snapshot` | varchar(160) | ชื่อตอนสั่ง |
| `customer_phone_snapshot` | varchar(40) | เบอร์ตอนสั่ง |
| `fulfillment_type` | varchar(20) | `delivery`, `pickup` |
| `delivery_address_snapshot` | jsonb nullable | ที่อยู่ตอนสั่ง |
| `delivery_note` | text nullable | หมายเหตุการจัดส่ง |
| `pickup_at` | timestamptz nullable | วันเวลารับสินค้า |
| `subtotal_satang` | bigint | ยอดก่อนค่าส่งและส่วนลด |
| `delivery_fee_satang` | bigint | ค่าส่ง |
| `discount_satang` | bigint | ส่วนลด default `0` |
| `total_satang` | bigint | ยอดสุทธิ |
| `currency` | char(3) | default `THB` |
| `status` | varchar(30) | สถานะออเดอร์ |
| `internal_note` | text nullable | เฉพาะ staff/admin |
| `placed_at` | timestamptz nullable | เวลาสั่งซื้อ |
| audit columns | — | ตามมาตรฐานด้านบน |

`delivery_address_snapshot` ควรเก็บชื่อผู้รับ ที่อยู่ จังหวัด รหัสไปรษณีย์ และเบอร์โทรที่ใช้จัดส่งจริง

### `order_items`

`id`, `order_id FK`, `product_id FK nullable`, `product_name_snapshot`, `unit_price_satang`, `quantity`, `line_total_satang`, `special_instructions`, audit columns

```sql
CHECK (quantity > 0),
CHECK (unit_price_satang >= 0),
CHECK (line_total_satang = unit_price_satang * quantity)
```

ออเดอร์เก่าต้องใช้ราคาและชื่อจาก snapshot ห้ามคำนวณใหม่จาก `products`

## 8. ตารางการชำระเงิน

### `payments`

หนึ่งออเดอร์สามารถมีหลาย payment เพื่อรองรับการส่งสลิปใหม่หรือเปลี่ยนวิธีชำระเงิน

`id`, `order_id FK`, `method`, `expected_amount_satang`, `paid_amount_satang`, `status`, `verified_by FK nullable`, `verified_at`, `rejection_reason`, `idempotency_key UNIQUE`, audit columns

สถานะ: `pending`, `submitted`, `approved`, `rejected`, `needs_resubmission`, `expired`

- `approved` ต้องมี `verified_by` และ `verified_at`
- `rejected` ต้องมี `rejection_reason`
- payment ที่อนุมัติต้องตรวจยอดกับ `expected_amount_satang`

### `payment_evidence`

`id`, `payment_id FK`, `object_key`, `mime_type`, `file_size_bytes`, `checksum`, `submitted_at`, audit columns

ใช้ signed URL และตรวจ MIME type/ขนาดไฟล์ที่ backend ก่อนอัปโหลด

## 9. ตารางสต็อก

### `inventory_movements`

`id`, `product_id FK`, `quantity_delta`, `reason`, `reference_type`, `reference_id`, `idempotency_key UNIQUE`, `created_by FK nullable`, `created_at`

`reason`: `restock`, `sale`, `cancellation`, `refund`, `adjustment`

### `inventory_reservations`

`id`, `order_id FK`, `product_id FK`, `quantity`, `status`, `expires_at`, `released_at`, audit columns

กำหนด `UNIQUE (order_id, product_id)` และ `CHECK (quantity > 0)` เพื่อให้ออเดอร์หนึ่งจองได้หลายสินค้า

`status`: `reserved`, `committed`, `released`, `expired`

ใช้ reservation กันสินค้าระหว่างรอชำระเงิน และตั้ง job สำหรับปล่อย reservation ที่หมดอายุ

## 10. สถานะออเดอร์

```text
pending_payment → pending_verification → confirmed → preparing
pending_payment → expired
pending_verification / confirmed / preparing → cancelled
preparing → ready_for_pickup / shipped → completed
```

ทุกการเปลี่ยนสถานะต้องเกิดใน transaction และสร้าง `order_status_history` เสมอ

### `order_status_history`

`id`, `order_id`, `from_status`, `to_status`, `changed_by`, `note`, `created_at`

ตารางนี้เป็น append-only

## 11. Audit log

### `audit_logs`

`id`, `actor_id`, `action`, `entity_type`, `entity_id`, `before_json`, `after_json`, `request_id`, `created_at`

ไม่ควรแก้ไขหรือลบ audit log จาก application ปกติ

## 12. Transaction สำคัญ

### สร้างออเดอร์

1. ตรวจ `idempotency_key` ถ้ามีออเดอร์เดิมให้คืนผลเดิม
2. Lock สินค้าที่เกี่ยวข้องด้วย `SELECT ... FOR UPDATE`
3. Lock สินค้าตามลำดับ `id` เดียวกันทุก transaction เพื่อลด deadlock และตรวจ `available = stock_quantity - SUM(quantity ของ reservation ที่ยัง reserved)`
4. สร้าง `orders` และ `order_items` พร้อม snapshot
5. สร้าง `inventory_reservations` พร้อม `expires_at`
6. สร้าง `payments` เป็น `pending`
7. สร้าง `order_status_history`
8. Commit แล้วจึงส่ง notification ผ่าน queue

หากขั้นตอนใดล้มเหลวต้อง rollback ทั้งชุด

### อนุมัติการชำระเงิน

Lock order/payment และสินค้าตามลำดับเดียวกัน ตรวจสถานะและ reservation ว่ายังใช้ได้ ตรวจว่าออเดอร์ยังไม่มี payment ที่อนุมัติแล้ว เปลี่ยน payment เป็น `approved` และ order เป็น `confirmed` ลด `stock_quantity` พร้อมสร้าง movement `sale` แล้วเปลี่ยน reservation เป็น `committed` ใน transaction เดียวกัน บันทึก history และ audit log ก่อน commit

หาก reservation หมดอายุ ให้ตรวจและจองสต็อกใหม่ก่อนอนุมัติ หากสินค้าไม่พอให้เข้าสู่ขั้นตอนจัดการด้วย staff ไม่ยืนยันออเดอร์อัตโนมัติ การคืนเงินไม่เท่ากับการคืนสินค้าเข้าสต็อก ต้องเติมสต็อกเฉพาะสินค้าที่ขายต่อได้

เมื่อยกเลิกออเดอร์ที่ยังไม่ได้ตัดสต็อก ให้ปล่อย reservation โดยไม่เพิ่ม `stock_quantity` เมื่อยกเลิกหลังตัดสต็อก ให้คืนสต็อกและสร้าง movement เพียงครั้งเดียวตามกฎธุรกิจ

## 13. Index ที่ควรมี

```sql
CREATE INDEX ix_products_category_active
ON app_schema.products(category_id, is_active, is_featured)
WHERE deleted_at IS NULL;

CREATE INDEX ix_orders_customer_created
ON app_schema.orders(customer_id, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX ix_orders_status_created
ON app_schema.orders(status, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX ix_payments_status_created
ON app_schema.payments(status, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX ix_inventory_product_created
ON app_schema.inventory_movements(product_id, created_at DESC);

CREATE INDEX ix_audit_entity_created
ON app_schema.audit_logs(entity_type, entity_id, created_at DESC);
```

## 14. Production checklist

- ใช้ migration tool เช่น Dbmate
- สร้าง `app_schema` ก่อน migration ตารางอื่น
- ทำ backup และ point-in-time recovery
- ใช้ TLS ระหว่าง backend กับ PostgreSQL
- แยก database role สำหรับ migration, application และ read-only reporting
- ใช้ signed URL สำหรับรูปและสลิป
- ตั้ง job สำหรับ expire reservation และออเดอร์ที่ไม่ชำระเงิน
- ทำ monitoring สำหรับ stock ติดลบ, payment ค้าง และ queue failure
- ทดสอบ transaction พร้อมกันหลาย request เพื่อป้องกัน overselling
