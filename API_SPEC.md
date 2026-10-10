# Olive & Oven — API ที่ควรใช้เมื่อเชื่อม Backend จริง

ปัจจุบันเดโมใช้ `localStorage` เท่านั้น เอกสารนี้เป็นสัญญา API เบื้องต้นสำหรับเปลี่ยน mock service เป็น REST API ภายหลัง

## มาตรฐานร่วม

- Base URL: `/api/v1`
- ทุก response เป็น JSON
- เงินเก็บเป็นจำนวนเต็มหน่วย THB เช่น `250`
- วันที่ใช้ ISO 8601 เช่น `2026-10-10T10:30:00+07:00`
- Admin endpoint ต้องตรวจสิทธิ์จาก backend เท่านั้น ห้ามเชื่อถือ localStorage

## Storefront

| Method | Endpoint | ใช้สำหรับ |
|---|---|---|
| GET | `/store` | โหลดข้อมูลร้าน เวลาเปิด banner delivery fee และ LINE OA URL |
| GET | `/categories` | โหลดหมวดหมู่สินค้า |
| GET | `/products?category=&q=&featured=&active=true` | ค้นหา/กรองสินค้า |
| GET | `/products/:productId` | รายละเอียดสินค้าและสินค้าที่เกี่ยวข้อง |
| GET | `/inventory/:productId` | เช็ก stock ล่าสุดก่อนเพิ่มตะกร้า |

## Cart และ Checkout

| Method | Endpoint | ใช้สำหรับ |
|---|---|---|
| POST | `/checkout/quote` | คำนวณ subtotal, delivery fee และ total |
| POST | `/orders` | สร้างคำสั่งซื้อ โดย backend snapshot ราคาและตัด stock แบบ atomic |
| GET | `/orders/:orderId` | อ่านรายละเอียดและสถานะคำสั่งซื้อ |
| GET | `/orders?customerId=` | ประวัติคำสั่งซื้อของลูกค้า |
| POST | `/orders/:orderId/cancel` | ขอ/อนุมัติการยกเลิก |

## Payment และ Slip

| Method | Endpoint | ใช้สำหรับ |
|---|---|---|
| GET | `/payments/:paymentId` | ดูยอดที่ต้องชำระและสถานะ |
| POST | `/payments/:paymentId/slip` | อัปโหลด slip แบบ multipart ไป object storage |
| POST | `/payments/:paymentId/submit` | แจ้งว่าลูกค้าส่งหลักฐานแล้ว |
| POST | `/admin/payments/:paymentId/approve` | Admin อนุมัติด้วยตนเอง |
| POST | `/admin/payments/:paymentId/reject` | Admin ปฏิเสธหรือขอ slip ใหม่ |

สถานะ: `pending`, `submitted`, `approved`, `rejected`, `needs_resubmission` — การอัปโหลดไฟล์ห้ามเปลี่ยนเป็น `approved` อัตโนมัติ

## Admin

| Method | Endpoint | ใช้สำหรับ |
|---|---|---|
| GET | `/admin/dashboard?from=&to=` | KPI, sales trend, status distribution |
| GET | `/admin/orders?status=&paymentStatus=&q=` | ตารางคำสั่งซื้อและตัวกรอง |
| PATCH | `/admin/orders/:orderId` | แก้รายการ delivery fee และ internal note |
| PATCH | `/admin/orders/:orderId/status` | เปลี่ยน order status พร้อม audit log |
| GET/POST/PATCH/DELETE | `/admin/products` | CRUD สินค้า |
| PATCH | `/admin/products/:productId/inventory` | ปรับ stock พร้อมเหตุผล |
| GET/PATCH | `/admin/settings` | อ่าน/แก้ store settings |
| GET | `/admin/audit-logs?entityType=&entityId=` | ประวัติการแก้ไข |

## LINE / LIFF integration layer

- `GET /integrations/line/config` — คืนเฉพาะ public LIFF ID และ OA URL
- `POST /integrations/line/identity` — แลก LIFF login token เป็น customer session
- `POST /integrations/line/orders/:orderId/notification` — ส่งแจ้งเตือนหลัง backend ตรวจสิทธิ์

ห้ามใส่ Channel Secret, access token หรือ credentials ใน frontend

## Error format

```json
{
  "error": { "code": "OUT_OF_STOCK", "message": "สินค้าไม่เพียงพอ", "fields": { "items[0].quantity": "เหลือ 1 ชิ้น" } },
  "requestId": "req_demo_123"
}
```
