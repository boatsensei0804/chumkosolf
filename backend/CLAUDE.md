# backend/CLAUDE.md — chumkosoft Golang Backend

อ่านไฟล์นี้ก่อนแตะโค้ด backend ทุกครั้ง ใช้คู่กับ `../CLAUDE.md` (context หลัก)

---

## 1. Stack

- **Go** + **Fiber**
- **PostgreSQL** ผ่าน **pgx** (+ sqlc สำหรับ type-safe query) — ถ้าเลือก GORM ให้ระบุเหตุผลและใช้ทั้งโปรเจคให้สม่ำเสมอ
- **Redis** สำหรับ session/refresh token, cache, rate limit
- **golang-migrate** สำหรับ migration
- **JWT** (access + refresh)
- **go-playground/validator** สำหรับ validate request

---

## 2. Clean Architecture (แยกชั้นเสมอ)

```
internal/
├── domain/        entities + business rules (ไม่พึ่ง framework)
├── repository/    database access (บังคับ school_id + semester_id ที่นี่)
├── service/       business logic
├── handler/       HTTP handlers (Fiber) — แปลง request/response เท่านั้น
└── middleware/    auth, tenant context, logging, recovery, CORS
```

- handler บาง: ทำแค่ parse + validate request → เรียก service → ตอบ response
- business logic อยู่ใน service ไม่ปนใน handler หรือ repository
- repository ไม่มี business logic แต่**ต้องบังคับ tenant + term scoping** (ดูข้อ 4)

---

## 3. Standard API

- RESTful + versioning ผ่าน URL: `/api/v1/...`
- Response format สม่ำเสมอ:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": { "page": 1, "total": 100 }
}
```

- ใช้ HTTP status code ให้ถูกต้อง
- เขียน OpenAPI/Swagger (swaggo) ให้ตรงกับ endpoint จริงเสมอ
- error response มีโครงสร้างชัด มี code + message ภาษาไทยที่ผู้ใช้เข้าใจได้

---

## 4. Tenant + Term Scoping (กฎที่สำคัญที่สุดของ backend)

ความเสี่ยงใหญ่ที่สุดคือข้อมูลรั่วข้ามโรงเรียน หรือข้อมูลปนข้ามเทอม ป้องกันด้วย:

1. middleware ดึง `school_id` (และ role, กลุ่มงาน) จาก JWT → ใส่ใน `context.Context`
2. `semester_id` ปัจจุบันมาจาก request หรือ resolve จากเทอมที่ active
3. **ทุก query ใน repository ต้อง filter `school_id`** และถ้าเป็นข้อมูลรายเทอมต้อง filter `semester_id` ด้วย
4. handler **ห้ามรับ `school_id` จาก client โดยตรง** — ดึงจาก context เท่านั้น
5. (แนะนำ) เปิด PostgreSQL Row-Level Security เป็นชั้นป้องกันสุดท้าย

ดูรายละเอียดและตัวอย่างใน skill `tenant-term-scoping`

---

## 5. Permission

บังคับสิทธิ์ที่ middleware/service ตามโมเดล 3 ชั้น (ดู context หลัก ข้อ 4.6):

- ตรวจ role (super_admin / teacher / executive / student)
- ตรวจ `is_school_admin`
- ตรวจการเป็นสมาชิก/admin ของกลุ่มงาน (`user_work_groups`)
- สิทธิ์ครูต่อ "วิชาตัวเอง" / "ห้องที่ปรึกษาตัวเอง" ตรวจจากความเป็นเจ้าของจริง ไม่ใช่แค่ role

---

## 6. การทดสอบ (บังคับหลังทำเสร็จทุกครั้ง)

ก่อนถือว่าเสร็จ ต้องผ่าน:

- `go test ./...` — unit + integration test ผ่าน
- `golangci-lint run` — ผ่าน
- `go vet ./...` — ผ่าน
- feature ใหม่ต้องมี test, โดยเฉพาะ **test เรื่อง isolation**: โรงเรียน A เข้าถึงข้อมูลโรงเรียน B ไม่ได้ และข้อมูลเทอมหนึ่งไม่ปนอีกเทอม

ถ้าข้อใดไม่ผ่าน ห้ามรายงานว่าเสร็จ

---

## 6.5 ข้อมูลส่วนบุคคล / PDPA

ระบบเก็บข้อมูลอ่อนไหว (เลขบัตรประชาชน, ข้อมูลเด็ก) ทุกครั้งที่แตะข้อมูลส่วนบุคคล:

- เลขบัตรประชาชนเข้ารหัส/จำกัดสิทธิ์ตอนเก็บ, ไม่ส่งเต็มกลับ frontend ถ้าไม่จำเป็น (mask)
- มี audit log การ view/update/delete ข้อมูลส่วนบุคคล
- จำกัดเข้าถึงตามกลุ่มงาน (บุคคล→ครู, วิชาการ→นักเรียน/ผู้ปกครอง)
- ไฟล์แนบเข้าถึงผ่าน signed URL หมดอายุ
- ไม่ log ข้อมูลอ่อนไหวลง log/error
- ดู skill sensitive-data-handling

## 7. Migration

- ทุกการเปลี่ยน schema ผ่าน migration เท่านั้น ห้ามแก้ DB ด้วยมือ
- มีทั้ง `up` และ `down` เสมอ
- ทุกตารางข้อมูลโรงเรียนมี `school_id`; ข้อมูลรายเทอมมี `semester_id`
- ใช้ UUID เป็น primary key, soft delete (`deleted_at`) สำหรับข้อมูลสำคัญ
- ตั้ง index บน `school_id`, `semester_id`, foreign key ที่ query บ่อย
- ดู skill `db-migration`

---

## 8. กฎการแก้โค้ด (ย้ำจาก context หลัก)

- ห้ามลบไฟล์โดยไม่จำเป็น
- ห้ามแก้โค้ดที่ไม่เกี่ยวข้อง (minimal diff)
- ห้าม reformat ทั้งไฟล์โดยไม่จำเป็น (ใช้ gofmt ปกติได้ แต่ไม่แก้ logic อื่น)
- เห็นจุดควรปรับนอกขอบเขต → แจ้งผู้ใช้ ไม่แก้เอง
