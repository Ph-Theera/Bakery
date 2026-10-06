# Bakery

Go Product API ตาม Clean Architecture: domain → ports → usecases; REST และ PostgreSQL อยู่ใน adapters และประกอบ dependencies ที่ `cmd/bakery/wire`.

## เริ่มใช้งานกับ Supabase

ต้องใช้ Go 1.25+ และ PostgreSQL ของ Supabase

1. เปิด Supabase Dashboard → Connect แล้วคัดลอก PostgreSQL URI ใช้ Direct connection ถ้าเครือข่ายรองรับ IPv6 หรือ Session pooler สำหรับ IPv4 ตาม [เอกสาร Supabase](https://supabase.com/docs/guides/database/connecting-to-postgres)
2. คัดลอก `.env.example` เป็น `.env` แล้วแก้ `DATABASE_URL` ด้วย URI ของโปรเจกต์ เพิ่ม `sslmode=require` และ percent-encode อักขระพิเศษในรหัสผ่าน ไฟล์นี้ถูก gitignore แล้ว อย่าใส่ URL จริงลง repository
3. รัน `make dbmate ARGS="up" DBMATE_TARGET=supabase` เพื่อ apply migrations (ต้องมี Docker). ใช้ PostgreSQL URI แบบ Direct หรือ Session pooler. ถ้า URI สำหรับ migration ต่างจาก API ให้เพิ่ม `MIGRATION_DATABASE_URL` ใน `.env`
4. รัน `make run` เพื่อเริ่ม API ในเครื่อง หรือ `make compose-supabase` เพื่อเริ่ม Docker พร้อม hot reload ที่เชื่อม Supabase

`make run`, `make dev`, `make migrate` และ `make test-integration` โหลด `.env` อัตโนมัติผ่าน shell; ใส่ค่าที่มีอักขระพิเศษใน single quotes เช่น `DATABASE_URL='postgresql://...'`. โปรแกรม Go เองยังรับค่าผ่าน environment. ต้อง percent-encode รหัสผ่านใน URI ก่อนเสมอ

ค่าเริ่มต้นฟังที่ `127.0.0.1:8080` ตั้ง `HTTP_ADDR` เพื่อเปลี่ยน address. ยังไม่มีระบบ authentication ของ HTTP API; หากเปิดให้เข้าจากภายนอกต้องเพิ่มการควบคุมสิทธิ์ก่อน. Database adapter ใช้ `pgxpool` ไม่ใช้ Supabase REST API หรือ anon key. การเชื่อมต่อเป็น database role ไม่ได้ใช้ JWT ของผู้ใช้. Migration เปิด RLS โดยไม่มี public policies; role ที่ backend ใช้ต้องมีสิทธิ์อ่าน/เขียนตามที่ต้องการ (URI ของ role postgres ใน Connect ใช้กับการตั้งต้นได้).

## API

| Method | Path | ผลลัพธ์ |
|---|---|---|
| POST | `/products` | สร้างสินค้า, 201 + Location |
| GET | `/products?limit=20&offset=0` | รายการสินค้าเรียง id, limit 1–100 |
| GET | `/products/{id}` | รายละเอียดสินค้า |
| PUT | `/products/{id}` | แทนที่ข้อมูลสินค้าทั้งชุด |
| DELETE | `/products/{id}` | ลบถาวร, 204 |

ตัวอย่างสร้างสินค้า:

```bash
curl -i http://127.0.0.1:8080/products \
  -H 'Content-Type: application/json' \
  -d '{"sku":"CROISSANT-001","name":"ครัวซองต์","price_satang":6500,"stock":20}'
```

POST/PUT ต้องส่งทั้งสี่ฟิลด์ ราคาเป็น integer หน่วยสตางค์ (6500 = 65 บาท), stock เป็น integer ไม่ติดลบ; SKU ไม่ซ้ำ, ยาวสูงสุด 64 ตัวอักษร และชื่อสูงสุด 200 ตัวอักษร. ID เป็น positive bigint สร้างโดย PostgreSQL. Error: 400 input ไม่ถูกต้อง, 404 ไม่พบสินค้า, 409 SKU ซ้ำ, 500 ข้อผิดพลาดภายในโดยไม่เปิดเผย DB error. ข้อมูลถูกเก็บถาวรในตาราง `public.products`.

## Docker และ Makefile

ใช้โครงสร้างเดียวกับ shipment-service: `Dockerfile.dev` สำหรับ Air hot reload และ `Dockerfile` สำหรับแพ็ก static binary ที่ได้จาก `make build` ลง distroless image โดยใช้ nonroot user. Compose development เริ่ม Bakery + PostgreSQL + dbmate เหมือน shipment-service; แยก Compose สำหรับ Supabase และ production

```bash
cp .env.example .env
# ใส่ DATABASE_URL เฉพาะเมื่อต้องการโหมด Supabase
make compose-local     # development, foreground; Ctrl+C เพื่อหยุด
make compose-up        # development, background
make compose-logs
make compose-down

make docker-build      # build Linux binary และ production image bakery:local
make compose-prod      # build และเริ่ม production container
make compose-prod-down
```

Compose local ใช้ PostgreSQL ใน Docker และไม่ต้องมี `.env`; Compose Supabase/production อ่าน `DATABASE_URL` จาก `.env` และส่งผ่าน environment ตอน runtime. ภายใน container ใช้ `HTTP_ADDR=0.0.0.0:8080` เพื่อให้ port mapping ทำงาน; ฝั่ง host เปิดที่ `127.0.0.1:8080`. เปลี่ยน host port ด้วย `HTTP_PORT` ใน `.env` โดยยังใช้ container port 8080. ไม่ใส่ secrets ลง image และ `.dockerignore` ยกเว้น `.env`, `.aws` และไฟล์ keys. Compose local รัน dbmate up อัตโนมัติ แล้วเริ่ม API หลัง migration สำเร็จ; Supabase/production ให้รัน `make dbmate ARGS="up" DBMATE_TARGET=supabase` ก่อนเริ่ม API

Development bind-mount source เพื่อให้ Air rebuild เมื่อแก้ Go files และเก็บ module/build cache ใน named volumes. Production ไม่มี source mount และ filesystem เป็น read-only. Dev กับ production ใช้ Compose project เดียวกัน ให้หยุดโหมดเดิมก่อนสลับโหมด

```bash
make help              # ดูคำสั่งทั้งหมด
make tools-install     # ติดตั้ง golangci-lint และ Air รุ่นที่ตรึงไว้
make run               # รันในเครื่องพร้อมโหลด .env
make dev               # รันในเครื่องพร้อม hot reload
make fmt               # gofmt และ goimports
make lint              # ตรวจ format, static analysis และกฎ dependency ของ architecture
make test              # unit tests; ปิด TEST_DATABASE_URL เพื่อไม่แตะ external DB
make vet
make build             # static Linux binary ใน build/bakery
make check             # lint + unit tests + build
```

เพิ่ม `$(go env GOPATH)/bin` ลง `PATH` เพื่อใช้ tools หลังติดตั้ง. Makefile ตั้ง Go toolchain เป็น `go1.26.8` ให้สอดคล้องกับ development image (Go 1.26); override ด้วย `GOTOOLCHAIN` ได้. ถ้าสร้าง production image สำหรับ architecture อื่น ระบุ `GOARCH` ให้ตรงกับ Docker target เช่น `make docker-build GOARCH=arm64` บนเครื่อง Docker ARM64. `build/bakery` ใช้รันใน Linux หรือ container; บน macOS/Windows ให้ใช้ `make run`

การตั้งค่า lint อยู่ใน `.golangci.yml` ใช้ configuration v2 ตาม [เอกสาร golangci-lint](https://golangci-lint.run/docs/configuration/file/) และเพิ่ม depguard เพื่อห้ามชั้น domain/ports/usecases import adapters หรือ infrastructure

Integration test ของ DB ใช้ `TEST_DATABASE_URL` ของฐานข้อมูลทดสอบที่ migration แล้วและมีตาราง products ว่าง ไม่ใช้ production เพราะ test สร้างและลบข้อมูล:

```bash
make test-integration
```

## Migrations แบบ shipment-service

ใช้ dbmate Docker image `ghcr.io/amacneil/dbmate:2.16.0` เช่นเดียวกับ shipment-service และ mount `db/` ไป `/data`. dbmate บันทึกเวอร์ชันใน `schema_migrations`, apply เฉพาะ migration ที่ยังไม่รัน และ export `db/schema.sql` ตาม [คู่มือ dbmate](https://github.com/amacneil/dbmate)

```bash
make dbmate ARGS="up"
make dbmate ARGS="status"
make dbmate ARGS="new add_product_description"
make dbmate ARGS="dump"
make migrate              # alias ของ ARGS="up"
```

ไฟล์ migration มีส่วน `-- migrate:up` และ `-- migrate:down`; dbmate จัดการ transaction ให้. `make dbmate ARGS="rollback"` ย้อน migration ล่าสุด สำหรับ migration สร้าง products การ rollback จะลบตารางและข้อมูลทั้งหมด

คำสั่ง `make dbmate` ใช้ service dbmate ของ Compose local เป็นค่าเริ่มต้นเสมอ. หากต้องการ Supabase ให้ระบุ `DBMATE_TARGET=supabase` เพื่อใช้ URL จาก `.env` ผ่าน Docker runner

ค่า `DBMATE_NETWORK` เริ่มต้นเป็น `--network=host` ตาม shipment-service. Override ได้ เช่น `make dbmate ARGS="up" DBMATE_NETWORK="--network=bakery-test"` สำหรับ DB ใน Docker network เดียวกัน. คำสั่งอ่าน `DATABASE_URL` จาก `.env` หรือ environment; ไม่ต้องติดตั้ง dbmate/psql ในเครื่อง

ถ้าเคยรันไฟล์ `001_products.up.sql` เดิมด้วย psql แล้ว ตาราง products จะมีอยู่แต่ยังไม่มีประวัติ dbmate: `up` จะ fail เพื่อไม่แก้หรือลบข้อมูลเดิมโดยอัตโนมัติ ต้องตรวจ schema และ baseline migration ที่ apply ไปแล้วก่อนใช้งาน dbmate กับฐานข้อมูลนั้น

## เปิด local stack แบบ shipment-service

```bash
make compose-up
# หรือ make compose-local เพื่อดู logs ใน foreground
make dbmate ARGS="status"
make compose-logs
make compose-down
```

ไม่ต้องตั้งค่า Supabase เพื่อเปิด local stack. API อยู่ที่ `http://localhost:8080/products` และ PostgreSQL อยู่ที่ `localhost:55432` (user/password: `user`/`password`, database: `bakery`). ใช้ port 55432 เพื่อไม่ชนกับ shipment-service ที่ใช้ 5432; override ด้วย `POSTGRES_PORT` ได้. ข้อมูลอยู่ใน volume `postgres-data` และคงอยู่เมื่อสั่ง down. dbmate จะมีสถานะ Exited (0) หลัง migration สำเร็จ ซึ่งเป็นปกติ

Supabase ใช้ `make compose-supabase` หลังใส่ `DATABASE_URL` ใน `.env`. หยุด local API ก่อน เพราะใช้ host port 8080 เหมือนกัน. Local Compose ใช้ URL ของ PostgreSQL local เสมอ แม้ `.env` จะมี Supabase URL; และ `make dbmate` ใช้ local เป็นค่าเริ่มต้นเช่นกัน. สำหรับ Supabase ใช้ `make dbmate ARGS="up" DBMATE_TARGET=supabase`
