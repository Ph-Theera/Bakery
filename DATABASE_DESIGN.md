# Olive & Oven — Database Design

## สถาปัตยกรรมที่แนะนำ

- **PostgreSQL**: ข้อมูลสินค้า ลูกค้า ออเดอร์ การชำระเงิน และสต็อก
- **S3 / Cloudflare R2**: รูปสินค้าและสลิปโอนเงิน
- **Redis (เสริม)**: cache, rate limit และ queue แจ้งเตือน
- **Backend API**: เป็น source of truth สำหรับราคา stock สิทธิ์ และสถานะการชำระเงิน

ไม่ควรใช้ localStorage เป็นฐานข้อมูลจริง และไม่ควรเก็บไฟล์ภาพเป็น base64 ใน PostgreSQL

## ความสัมพันธ์หลัก

```text
users ──< customers ──< orders ──< order_items >── products
                         │              │
                         ├── payments ──< payment_evidence
                         ├── order_status_history
                         └── audit_logs

categories ──< products ──< product_images
products ──< inventory_movements
```

## ตารางหลัก

### users

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | รหัสผู้ใช้ |
| email | varchar unique nullable | สำหรับ admin login ในอนาคต |
| role | enum | `customer`, `admin`, `staff` |
| auth_provider | varchar nullable | `line`, `email`, `google` |
| provider_subject | varchar nullable | ห้ามเก็บ access token |
| is_active | boolean | สถานะบัญชี |
| created_at, updated_at | timestamptz | |

### customers

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| user_id | uuid FK nullable | รองรับ LINE/บัญชีในอนาคต |
| display_name | varchar | ชื่อลูกค้า |
| phone | varchar | ควรเข้ารหัสหรือ mask ตาม policy |
| line_display_name | varchar nullable | |
| created_at, updated_at | timestamptz | |

### categories

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| slug | varchar unique | เช่น `cookies` |
| name_th | varchar | เช่น คุกกี้ |
| sort_order | integer | ลำดับการแสดงผล |
| is_active | boolean | |

### products

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| sku | varchar unique | รหัสสินค้า |
| category_id | uuid FK | |
| name | varchar | ชื่อสินค้า |
| slug | varchar unique | สำหรับ URL |
| description_th | text | |
| price_thb | integer | ราคาปัจจุบัน หน่วยบาท |
| compare_at_price_thb | integer nullable | ราคาเดิม |
| stock_quantity | integer | stock ปัจจุบันแบบอ่านเร็ว |
| is_active | boolean | ซ่อนสินค้าเมื่อ false |
| is_featured | boolean | สินค้าแนะนำ |
| created_at, updated_at | timestamptz | |

Constraint: `price_thb >= 0`, `stock_quantity >= 0` และสินค้าที่เคยขายแล้วควร deactivate แทนการลบ

### product_images

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| product_id | uuid FK | |
| object_key | varchar | path ใน S3/R2 |
| public_url | text | URL ผ่าน CDN หรือ signed URL |
| alt_text | varchar | accessibility |
| sort_order | integer | |
| is_primary | boolean | รูปหลัก |

### orders

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | internal ID |
| order_number | varchar unique | เช่น `OO-260910-184` |
| customer_id | uuid FK nullable | รองรับ guest order |
| customer_name_snapshot | varchar | ชื่อตอนสั่ง |
| customer_phone_snapshot | varchar | เบอร์ snapshot |
| fulfillment_type | enum | `delivery`, `pickup` |
| delivery_address | jsonb nullable | ที่อยู่ snapshot |
| delivery_note | text nullable | |
| pickup_date, pickup_time | date/time nullable | |
| subtotal_thb | integer | snapshot ตอนสร้าง |
| delivery_fee_thb | integer | snapshot |
| discount_thb | integer | |
| total_thb | integer | ยอดสุทธิ |
| order_status | enum | สถานะออเดอร์ |
| internal_note | text nullable | เฉพาะ admin |
| placed_at, created_at, updated_at | timestamptz | |

### order_items

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| order_id | uuid FK | |
| product_id | uuid FK nullable | สินค้าอาจถูก archive ภายหลัง |
| product_name_snapshot | varchar | ชื่อตอนซื้อ |
| unit_price_thb | integer | ราคาต่อชิ้นตอนซื้อ |
| quantity | integer | มากกว่า 0 |
| line_total_thb | integer | `unit_price_thb * quantity` |
| special_instructions | text nullable | |

> สำคัญ: ห้ามคำนวณออเดอร์เก่าจากราคาปัจจุบันของ products ให้ใช้ snapshot ใน order_items

### payments

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| order_id | uuid FK unique | เวอร์ชันแรก 1 payment ต่อ order |
| method | enum | `bank_transfer`, `promptpay` |
| expected_amount_thb | integer | ยอดที่ต้องชำระ |
| status | enum | `pending`, `submitted`, `approved`, `rejected`, `needs_resubmission` |
| verified_by | uuid FK nullable | admin ผู้ตรวจ |
| verified_at | timestamptz nullable | |
| rejection_reason | text nullable | |
| created_at, updated_at | timestamptz | |

### payment_evidence

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| payment_id | uuid FK | |
| object_key | varchar | path ของไฟล์ใน S3/R2 |
| mime_type | varchar | JPG/PNG/WebP |
| file_size_bytes | integer | จำกัดขนาดไฟล์ |
| checksum | varchar | ตรวจ integrity |
| submitted_at | timestamptz | |

### inventory_movements

เก็บทุกการเปลี่ยนสต็อก ไม่แก้ตัวเลขโดยไม่มีประวัติ

| Column | Type | รายละเอียด |
|---|---|---|
| id | uuid PK | |
| product_id | uuid FK | |
| quantity_delta | integer | `-2` ขาย, `+10` เติมสต็อก |
| reason | enum | `sale`, `restock`, `adjustment`, `cancellation`, `refund` |
| reference_type, reference_id | varchar nullable | ผูกกับ order/admin action |
| created_by | uuid FK nullable | |
| created_at | timestamptz | |

### order_status_history และ audit_logs

- `order_status_history`: `order_id`, `from_status`, `to_status`, `changed_by`, `note`, `created_at`
- `audit_logs`: `actor_id`, `action`, `entity_type`, `entity_id`, `before_json`, `after_json`, `request_id`, `created_at`

## สถานะคำสั่งซื้อ

```text
pending_payment → pending_verification → confirmed → preparing
                                      → ready_for_pickup / shipped → completed

pending_payment / pending_verification / confirmed → cancelled
```

ทุกการเปลี่ยนสถานะต้องสร้าง `order_status_history`

## Transaction ตอนสร้างออเดอร์

1. Lock สินค้าด้วย `SELECT ... FOR UPDATE`
2. ตรวจ stock และราคาจาก backend
3. สร้าง order และ order_items พร้อม snapshot ราคา
4. ลด stock และสร้าง inventory movement
5. สร้าง payment เป็น `pending`
6. Commit แล้วจึงส่ง notification/LINE ผ่าน queue

ถ้าขั้นตอนใดล้มเหลวต้อง rollback ทั้งชุด เพื่อป้องกันออเดอร์เกิดแต่ stock ไม่ลด หรือ stock ติดลบ

## Index ที่ควรมี

- `products(category_id, is_active, is_featured)`
- `orders(customer_id, created_at desc)`
- `orders(order_status, created_at desc)`
- `orders(order_number)` unique
- `payments(status, created_at desc)`
- `inventory_movements(product_id, created_at desc)`
- `audit_logs(entity_type, entity_id, created_at desc)`

## ก่อนขึ้น Production

- ใช้ migration tool และ backup แบบ point-in-time recovery
- Encrypt connection และข้อมูลส่วนบุคคลที่จำเป็น
- ใช้ RBAC จาก backend เท่านั้น
- ใช้ signed URL สำหรับไฟล์สลิป
- เพิ่ม idempotency key ให้การสร้าง order และ submit payment
- ใช้ queue สำหรับ LINE notification
