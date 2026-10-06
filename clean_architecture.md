---
name: go-clean-architecture
description: สร้างหรือขยาย Go backend ด้วย Clean Architecture แบบ Ports & Adapters ตามรูปแบบ shipment-service ใช้เมื่อวางโครงสร้าง service ใหม่ เพิ่ม endpoint หรือแยก business logic ออกจาก HTTP และ persistence
---

# Go Clean Architecture

สร้าง Go service ที่แยก business logic, application workflow และระบบภายนอกด้วย ports และ dependency injection ใช้รูปแบบของ shipment-service เป็นแนวทาง และรักษา convention ของ repository ที่กำลังทำงาน

## ก่อนเริ่ม

- อ่าน `AGENTS.md`, เอกสาร architecture, `go.mod`, entrypoint และตัวอย่าง feature ที่ใกล้เคียง
- ระบุว่าเป็น service ใหม่หรือการเพิ่ม feature ใน service เดิม รวมถึง transport, database และ code generation ที่ใช้จริง
- ใน service เดิม ให้ต่อยอด package, constructor, error และ transaction convention เดิม หลีกเลี่ยงการเปลี่ยน architecture ทั้งระบบนอกขอบเขตงาน
- ใน service ใหม่ ให้สร้างเฉพาะ layer ที่ feature ต้องใช้ ไม่สร้าง interface, service หรือ integration เปล่าไว้ล่วงหน้า
- Gin, Bob, PostgreSQL, Supabase และ common-lib เป็นตัวเลือกของ shipment-service ไม่ใช่ข้อบังคับของ architecture

## โครงสร้างเป้าหมาย

```text
cmd/<service>/
  main.go
  wire/
    application.go
    rest_usecases.go
    dep/
internal/
  domains/<entity>/
  ports/
    port_in/
    port_out/
  usecases/
  services/
  adapters/
    inbound/rest/
    outbound/db/<entity>/
    outbound/auth/
    outbound/config/
    specs/
infrastructure/config/
pkg/
db/migrations/
```

ปรับชื่อ transport และ integration ตามงานจริง ไม่จำเป็นต้องสร้างทุกโฟลเดอร์

## กฎ dependency

ทิศทาง import หลัก:

```text
port_in  → domains
port_out → domains
usecases → port_in, port_out, domains, services
inbound adapters  → port_in, transport specifications
outbound adapters → port_out, external SDK / ORM / driver
wire → adapters, usecases, services, infrastructure
```

- `domains` ห้าม import ports, usecases, services, adapters หรือ infrastructure
- `ports`, `usecases` และ `services` ห้าม import adapters หรือ infrastructure
- HTTP request/response, ORM model และ vendor SDK response ต้องอยู่ฝั่ง adapter และแปลงก่อนข้าม boundary
- `wire` เป็น composition root จึง import concrete implementations ได้
- Runtime flow อาจเรียกจาก usecase ไป DB adapter แต่ usecase ต้องอ้างอิง repository interface เท่านั้น
- แยก architectural dependency ออกจาก utility dependency: UUID, decimal และ nullable types ใช้ใน domain ได้ตาม convention แต่ทำให้ domain มี third-party dependencies ซึ่งต้องระบุอย่างตรงไปตรงมา

## ความรับผิดชอบของแต่ละ layer

### Domains

เก็บ entity, value type, enum และ business error แยก package ตามแนวคิดธุรกิจ เช่น `shipment`, `company`, `container`

- วางกฎที่เป็นจริงเสมอของ entity และการเปลี่ยนสถานะไว้ใน domain เมื่อมีกฎนั้นจริง
- ใช้ `Data` สำหรับข้อมูลอ่าน และ `Setter` สำหรับ patch ได้ตามรูปแบบเดิม
- แยก omitted, null และ zero value ให้ชัดเมื่อ API รองรับ partial update
- ไม่ใส่ HTTP status, SQL query หรือการอ่าน global config ใน domain
- CRUD ที่ยังไม่มีกฎซับซ้อนใช้ data struct ได้ ไม่เพิ่ม behavior ที่ไม่มี requirement

### Input ports

ประกาศ usecase interface, Command/Query และผลลัพธ์ใน `port_in`

```go
type CreateShipmentUsecase interface {
    Handle(context.Context, CreateShipmentCommand) (*ShipmentData, error)
}
```

Command/Query เป็นข้อมูล application ไม่รับ Gin context หรือ generated API request โดยตรง ใช้ type alias ของ domain ได้เมื่อข้อมูลและความหมายตรงกัน หาก output ต้อง reshape หรือจำกัดข้อมูล ให้สร้าง application DTO

### Usecases

หนึ่ง operation ต่อไฟล์ใน package `usecases` ใช้ constructor รับ dependencies และคืน input port interface

- ตรวจ input และประสานขั้นตอนธุรกิจ รวมถึง transaction boundary
- เรียก repository/client/provider ผ่าน `port_out`
- เรียก domain behavior สำหรับกฎของ entity
- ส่ง business error กลับโดยไม่กำหนด HTTP response
- ฉีดเฉพาะ dependencies ที่ operation ใช้จริง

### Shared services

เพิ่ม service เมื่อมี logic ที่ใช้ซ้ำหรือมีความรับผิดชอบชัดเจน

- Pure domain logic ใช้ domain types และไม่เรียก repository หรือคืน input-port DTO
- Service ที่เรียก repository หรือรวมข้อมูลหลาย query เป็น application service วางใน `internal/services` ได้ และใช้ ports
- ในรูปแบบเดิม `ShipmentService` มีทั้ง normalization และ query orchestration ให้ระบุว่าเป็น application service หากเพิ่ม behavior ในส่วนนี้
- ไม่สร้าง service ที่เพียงส่งต่อทุก method ของ repository โดยไม่มี logic เพิ่ม

### Output ports

ประกาศ repository, auth client, config provider หรือ integration contract ที่ application ต้องใช้ใน `port_out`

- ใช้ domain/application types ใน parameters และผลลัพธ์
- จำกัด method ตามความต้องการของ usecase ไม่เปิด vendor SDK ทั้งชุดผ่าน interface
- ใช้ config provider เมื่อ business workflow ต้องใช้ configuration ไม่อ่าน `config.Current` ใน usecase

### Adapters

Inbound REST handler แปลง request เป็น Command/Query เรียก input port แล้วแปลงผลลัพธ์และ error เป็น response กฎธุรกิจต้องอยู่ด้านใน

Outbound DB adapter เก็บ ORM/SQL, filter query และ mapper สำหรับแปลง DB model ↔ domain/output-port types ส่วน auth adapter ซ่อน vendor SDK ไว้หลัง auth port

Generated API และ ORM code อยู่ฝั่ง adapters แก้ source specification/schema และ regenerate แทนการแก้ generated files ตรวจผล diff ทุกครั้ง และไม่สั่ง generator ที่ลบ output ทั้ง directory โดยไม่ได้ตรวจผลกระทบก่อน

### Bootstrap และ wiring

`main.go` โหลด config และเริ่ม lifecycle ส่วน `wire/application.go` สร้าง adapters → shared services → usecases และส่ง input ports ให้ inbound server

อ่าน global config ได้ที่ infrastructure/bootstrap แล้วส่งค่าหรือ provider เข้า application หลีกเลี่ยงการประกอบ dependencies ภายใน handler/usecase

## Transaction และ framework coupling

เลือกรูปแบบให้ตรงกับบริบท:

- **Service เดิมแบบ shipment-service:** repositories รับ `*sql.Tx` และ usecase ใช้ `ExecWithTx` จาก common-lib ให้รักษา contract เดิมในการเพิ่ม feature และระบุว่า application ยังผูกกับ SQL/framework
- **Service ใหม่ที่ต้องการแยกเข้มขึ้น:** ประกาศ transaction/unit-of-work port โดยไม่เปิดเผย `*sql.Tx` ให้ callback ได้ repository scope ที่ใช้ transaction เดียวกัน และให้ DB adapter จัดการ begin/commit/rollback

Operation ที่เขียนหลาย repository แบบ atomic ต้องใช้ transaction เดียวกันทั้งหมด เมื่อ callback ล้มเหลวต้อง rollback และเมื่อสำเร็จต้อง commit รวมถึงส่ง commit error กลับ ไม่กลืน error

ใน service ใหม่ ฉีด Clock/IDGenerator เมื่อ application ต้องควบคุมเวลา/ID เพื่อทดสอบ ใช้ utility เดิมใน service ที่มี convention อยู่แล้ว โดยไม่ refactor นอก scope

## Error boundary

- Domain เป็นแหล่งนิยาม business errors เช่น not found, invalid command, invalid reference และ duplicate
- Ports สามารถ re-export domain errors เพื่อให้ caller ใช้ contract ของ layer ตัวเอง
- DB adapter แปลง driver errors เป็น business errors เช่น no rows → not found และ unique violation → duplicate
- Usecase propagate หรือ wrap ด้วย `%w` เพื่อให้ `errors.Is` ทำงาน
- REST adapter map เป็น HTTP status ตาม API contract ไม่ส่ง raw SQL/SDK error หรือข้อมูลลับให้ client

## ขั้นตอนเพิ่ม feature

1. ระบุ input, output, business rules และ external operations ที่ต้องใช้
2. เพิ่มหรือปรับ domain types/behavior เท่าที่จำเป็น
3. เพิ่ม input port และ output port ที่ยังขาด
4. เขียน usecase พร้อม dependencies ผ่าน constructor
5. Implement outbound adapter และ mapper ตาม contract
6. ปรับ API specification และ generate ตาม workflow ของ repository แล้วเขียน inbound handler
7. ต่อ dependencies ที่ composition root และลงทะเบียน operation กับ server
8. สร้างหรืออัปเดต mocks ตาม workflow เดิม และทดสอบส่วนที่เปลี่ยน

หากต้องเปลี่ยน schema ให้เพิ่ม migration และอัปเดต schema/generated models ตามเครื่องมือเดิม ไม่ใช้ migration หรือ generation ที่ต้องต่อ live database โดยไม่ตรวจ environment ก่อน

## การตรวจงาน

- ตรวจ imports ว่าไม่มี dependency จากชั้นในไป adapters/infrastructure
- ทดสอบ usecase ด้วย mocked output ports: success, business failure และการ propagate error ที่เกี่ยวข้อง
- ทดสอบ domain behavior เมื่อเพิ่มกฎธุรกิจ
- ทดสอบ adapter mapping/query หรือ HTTP error mapping ที่เปลี่ยน โดยเฉพาะ null/omitted และ soft delete
- สำหรับหลาย writes ใน transaction ตรวจ failure ของ write ถัดไป, rollback และ commit failure ตามขอบเขต implementation ที่เปลี่ยน
- ใช้ test framework และ build/lint commands ของ repository เดิม ไม่บังคับ Ginkgo หรือ coverage threshold กับทุกโปรเจกต์
- สรุป layer ที่เพิ่ม, dependency ที่ต่อ, checks ที่รัน และ coupling/ข้อจำกัดที่ยังเหลือ

งานเสร็จเมื่อ feature ต่อใช้งานได้จริงผ่าน wiring, checks ที่เกี่ยวข้องผ่าน และไม่มี transport/ORM types หลุดเข้ามาใน application โดยไม่ตั้งใจ
