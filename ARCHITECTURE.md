# TongBai API architecture

จัดโครงสร้างตามแนวทางของ shipment-service: แยก input/output ports, usecase ราย operation ที่มี constructor รับ Args และคืน input port และแยก infrastructure wiring ออกจาก application wiring

```text
cmd/bakery/wire/
  dep/dependencies.go          database และ startup configuration
  application.go              ประกอบ repositories, services และ usecases
  rest_usecases.go             ส่ง input ports ให้ REST
  server.go                   HTTP wiring และ LINE worker lifecycle
internal/
  shared/                     errors, UUID, opaque transaction contracts
  domains/                    catalog, orders, users, store, notifications
  ports/
    port_in/                  interface + Command/Query ราย operation
    port_out/                 repository, identity, sender และ outbox contracts
  usecases/
    product/                  CRUD catalog และ stock-update policy
    users/                    customers และ verified LINE identity
    orders/                   queries, access, notes และ dashboard
    payments/                 payment queries
    store/                    settings
    composite/                checkout, order/payment workflows, notifications
  services/checkout/          pricing ผ่าน catalog/inventory/store ports
  adapters/
    inbound/rest/v1/           routes, auth, HTTP decoding/encoding
    outbound/db/              catalog, users, orders, payments, inventory,
                              store, notifications และ transaction executor
    outbound/line/            verify identity และส่งข้อความ
```

## Dependency rules

shared เป็น dependency ล่างสุด ใช้เฉพาะ standard library และไม่ import feature ใด ๆ Domain เป็นเจ้าของข้อมูลและ business rules เช่น order transitions และ payment-review policy

REST เรียกผ่าน port_in เท่านั้น ไม่สร้าง usecase หรือเรียก repository เอง Usecase implement input port และเรียก output ports; DB/LINE adapters implement output ports โดยไม่ import usecases กฎเหล่านี้ตรวจด้วย depguard ใน .golangci.yml

Product และ users แยก folder ของตัวเอง Composite ใช้ output ports จากหลาย feature เมื่อ workflow ต้องทำงานร่วมกัน เช่น create order ใช้ catalog, inventory, users, orders, payments และ store ไม่ย้าย feature repositories ลง shared

## Transactions

Usecase เป็นเจ้าของ transaction boundary ผ่าน shared.ExecWithTx และส่ง transaction scope เดียวกันให้ทุก repository ใน callback ตัว DB adapter เป็นผู้ถือ pgx.Tx จริง โดยไม่มี driver type หลุดเข้า domain/ports/usecases

Payment approval ทำ stock deduction, inventory movements, reservation commit, payment/order status, history และ LINE outbox ใน transaction เดียวกัน หากขั้นใดผิดพลาดจะ rollback ทั้งชุด การอนุมัติซ้ำไม่ตัด stock หรือเพิ่ม outbox ซ้ำ และ LINE worker ส่งข้อความหลัง commit เท่านั้น

DB repositories ทำ SQL primitives ส่วนการตรวจ order/payment state และการประสานหลาย repositories อยู่ใน domain/usecase ไม่ใช่ aggregate repository ก้อนเดียว

## เทียบกับ shipment-service

ใช้รูปแบบ constructor Args → input port → Handle(ctx, Command/Query), output repositories และ wiring แยก dep, application, rest_usecases ตาม reference แต่ยังคง net/http และ pgx ของ Bakery ไม่เปลี่ยนเป็น Gin/Bob หรือเพิ่ม private common-lib จาก shipment-service

Reference ส่ง *sql.Tx ให้ repositories ส่วน Bakery ใช้ opaque transaction contract เพื่อให้ shared ไม่ขึ้นกับ database driver ทั้งสองแบบแชร์ transaction เดียวกันตลอด workflow

## Compatibility และ verification

Router ยังอยู่ที่ /api/v1 พร้อม unversioned aliases และคง request/response contract เดิม โค้ด product แบบ bigint รุ่นเดิมยังเก็บไว้เพื่อ review แต่ไม่ได้ wire เข้า UUID API ใหม่ ชื่อ domain/ports catalog ช่วยไม่ให้ทับกับ legacy product/products

ไฟล์ชุด API เดิมถูกย้าย/แยกเป็น ports, domains, usecases และ adapters ข้างต้น โดยไม่เปลี่ยน database schema ดูเส้น API และวิธีทดสอบใน [API.md](API.md)

Tests ครอบคลุม checkout replay, concurrent reservations, duplicate approvals, slips, cancellation, stock/outbox rollback และ transaction commit error ใช้ PostgreSQL database สำหรับ test ที่แยกจาก database ใช้งานจริงเท่านั้น
