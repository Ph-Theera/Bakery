# TongBai — Production Frontend

Next.js App Router frontend สำหรับเชื่อมต่อ REST API จริงของ TongBai

## Run

```powershell
cd web
copy .env.example .env.local
npm install
npm run dev
```

ตั้งค่า `NEXT_PUBLIC_API_BASE_URL` ให้ชี้ไปที่ backend เช่น `http://localhost:8080/api/v1`. ค่าเริ่มต้น `NEXT_PUBLIC_API_MOCK=false` ไม่ซ่อน API error; ตั้ง true เฉพาะ demo (ไม่ส่งคำขอจริง).

Backend ต้องรันที่ port 8080, apply migration API แล้ว และตั้ง `AUTH_SECRET` อย่างน้อย 32 ตัวอักษร. ตั้ง `FRONTEND_ORIGIN=http://localhost:3000` ให้ตรง origin เว็บจริงเพื่อให้ CORS อนุญาต Authorization และ Idempotency-Key. ไม่มีการ apply migration ฐานจริงจาก frontend.

หน้า `/admin` ให้กรอก `ADMIN_API_TOKEN` ของ backend (ซึ่งตรวจ `ADMIN_USER_ID` เป็น active admin). Token เก็บเฉพาะใน memory ไม่ใช้ NEXT_PUBLIC env/localStorage; refresh ต้องกรอกใหม่. เป็น bootstrap admin access ไม่ใช่ระบบ staff login แบบเต็ม ต้องใช้ HTTPS ใน production.

Checkout ส่ง Idempotency-Key เดิมเมื่อ retry payload เดิม; order token เก็บใน sessionStorage ของแท็บเพื่อส่งสลิปและติดตามสถานะ. ไม่มี server session. ปิดแท็บ/ล้าง storage จะเสีย guest access; LINE OAuth/LIFF login UI ยังไม่ได้ตั้งค่า แม้ client รองรับ `api.linkLine(accessToken)` แล้ว.

สร้างออเดอร์ก่อน จากนั้นอัปโหลดสลิปได้ทั้งโอนธนาคารและ PromptPay. ร้านอนุมัติ/ปฏิเสธผ่านหน้า admin; backend เป็นผู้ตัด stock และ enqueue LINE. ไม่มี QR หรือเลขบัญชีตัวอย่างแสดงเป็นข้อมูลจริง ต้องเพิ่ม payment instructions จากร้านก่อนเปิดรับเงินจริง. API ยังไม่รองรับอัปโหลดรูปสินค้า.

ตรวจด้วย `npm.cmd test`, `npm.cmd run typecheck`, `npm.cmd run build` (Windows PowerShell). Dev ใช้ .next-dev และ production build/start ใช้ .next-build เพื่อไม่ให้ cache ชนกัน.

`tests/integration.test.cjs` ไม่รันโดยปกติ ต้องเปิด `TEST_LIVE_API=true` เฉพาะเมื่อ Go API ต่อฐาน disposable ที่ seed สินค้า `integration-bread` แล้ว. Test เขียนออเดอร์/อนุมัติเงิน/แก้ settings และสินค้า ห้ามใช้กับร้านหรือฐานข้อมูลจริง. ทดสอบชุดนี้กับ PostgreSQL ชั่วคราวผ่านแล้ว และเก็บกวาด container หลังทดสอบ.

## Routes

- `/` storefront
- `/products/[slug]` product detail
- `/checkout` checkout และสร้าง order ผ่าน API
- `/orders` order history
- `/admin` admin dashboard

Frontend จะไม่ถือ localStorage เป็น source of truth: localStorage ใช้เพียงเก็บ cart UX ชั่วคราว ส่วนราคา stock order payment และ permission ต้องมาจาก backend
