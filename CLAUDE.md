# CLAUDE.md — chumkosoft (School Management System)

เอกสารนี้คือ context และกฎหลักของโปรเจค **ทุก AI agent / ทุก session ต้องอ่านและทำตามไฟล์นี้เสมอ** ก่อนเขียนหรือแก้โค้ดใด ๆ

---

## 1. โปรเจคนี้คืออะไร

**chumkosoft** — ระบบบริหารจัดการโรงเรียน สำหรับโรงเรียนระดับมัธยม/ประถมในประเทศไทย

ขอบเขตหลัก: จัดการข้อมูลบุคลากร (ครู/ผู้บริหาร), นักเรียน/ผู้ปกครอง, ห้องเรียน, ตารางสอน, การเช็คชื่อ และคะแนนความประพฤติ โดยแบ่งความรับผิดชอบตาม "กลุ่มงาน" ของโรงเรียน

### กลุ่มงาน (Work Groups)
1. **กลุ่มงานบุคคล** (`personnel`) — จัดการข้อมูลครู/บุคลากร
2. **กลุ่มงานบริหารทั่วไป** (`general_affairs`) — กิจการนักเรียน, คะแนนความประพฤติ, เช็คชื่อเข้าเรียน, กิจกรรม
3. **กลุ่มงานวิชาการ** (`academic`) — ข้อมูลนักเรียน, ผู้ปกครอง, เกรด, ตารางเรียน/สอน
4. **กลุ่มงานงบประมาณและแผน** (`budget_plan`) — งบประมาณ, โครงการ, เบิกจ่าย, พัสดุ (เฟสหลัง)

---

## 2. Tech Stack (ห้ามเปลี่ยนโดยไม่ปรึกษา)

### Backend
- **ภาษา:** Go (Golang) + **Fiber**
- **Database:** PostgreSQL ผ่าน pgx (+ sqlc) — ดู backend/CLAUDE.md
- **Cache/Session:** Redis
- **Migration:** golang-migrate
- **Auth:** JWT (access + refresh token)
- **Validation:** go-playground/validator

### Frontend
- **Framework:** Next.js (App Router) + **TypeScript** (strict, ห้าม any)
- **UI Library:** Ant Design (antd)
- **Styling:** Tailwind CSS
- **Form:** react-hook-form + **zod** (validation + type inference)
- **พาเลตสีหลัก:** ฟ้า-ขาว (ดู frontend/CLAUDE.md)

### Infra
- ทุก service รันใน **Docker container** (docker-compose)
- services: `frontend`, `backend`, `postgres`, `redis`
- ดูรายละเอียดในส่วน Docker ด้านล่าง

---

## 3. กฎเหล็กในการแก้ไขโค้ด (อ่านให้ดี — สำคัญที่สุด)

### 3.1 ห้ามลบไฟล์โดยไม่จำเป็น
อย่าลบไฟล์ใด ๆ เว้นแต่งานระบุชัดว่าให้ลบ หรือเป็นไฟล์ชั่วคราวที่ตัวเองเพิ่งสร้าง ถ้าไม่แน่ใจว่ายังถูกใช้อยู่ไหม ให้ grep ทั้ง repo ก่อนแล้วถามผู้ใช้ ห้ามลบเพื่อ "ทำความสะอาด" เอง

### 3.2 ห้ามแก้โค้ดที่ไม่เกี่ยวข้อง
แก้เฉพาะส่วนที่เกี่ยวกับงานที่ได้รับเท่านั้น ห้าม reformat ทั้งไฟล์, เปลี่ยน import ที่ไม่เกี่ยว, rename นอกขอบเขต, หรือ "ปรับปรุง" โค้ดอื่นที่ไม่ได้ถูกขอ เห็นจุดควรปรับนอกขอบเขต → บันทึกแล้วแจ้งผู้ใช้ ไม่แก้เอง ต้องการ minimal diff เสมอ

### 3.3 ทดสอบทุกครั้งหลังทำเสร็จ
หลังเขียน/แก้โค้ดทุกครั้ง ต้องรัน test และผ่านก่อนถือว่าเสร็จ
- Backend: `go test ./...` + `golangci-lint run` + `go vet ./...`
- Frontend: `npm run test` + `npm run lint` + `npm run type-check`
- feature ใหม่ต้องเขียน test ครอบคลุม ถ้า test ไม่ผ่าน ห้ามรายงานว่าเสร็จ

### 3.4 ทุกอย่างต้องอยู่ใน Business Logic ที่กำหนด
ห้ามสร้าง feature, endpoint, field ที่อยู่นอกขอบเขตในเอกสารนี้ ถ้างานดูเหมือนต้องการสิ่งนอกขอบเขต ให้หยุดถามผู้ใช้ก่อน อย่าเดาเอง

---

## 4. หลักการ Business Logic ที่ห้ามละเมิด

### 4.1 Multi-tenant เผื่อไว้ (school_id)
ทุกตารางข้อมูลโรงเรียนต้องมี `school_id` และทุก query ต้อง filter ด้วยค่านี้เสมอ ห้ามลบออก (เผื่อขยายหลายโรงเรียน) ดู skill tenant-term-scoping

### 4.2 ทุกข้อมูลผูกกับปีการศึกษา + เทอม (semester_id)
ข้อมูลที่เปลี่ยนตามเทอมต้องมี `semester_id` เสมอ
- ถาวร (ตัวบุคคล, รายชื่อวิชา) — ไม่ผูกเทอม
- รายเทอม (จัดห้อง, ลงทะเบียน, ตารางสอน, เช็คชื่อ, คะแนน, ผลงานครู) — ผูกเทอม

### 4.3 บุคลากร (กลุ่มงานบุคคล)
- ทุกคน (ครู/ผู้บริหาร/นักเรียน) login ผ่านตาราง `users` ร่วมกัน (เก็บ password_hash ที่นี่ ไม่ใช่ในตารางโปรไฟล์)
- **ครูและผู้บริหารใช้ตารางโปรไฟล์ `personnel` เดียวกัน** (ผอ./รอง ผอ. เป็นบุคลากรเหมือนครู มีวิทยฐานะ/สอนได้)
- ข้อมูล personnel: เลขบัตรประชาชน, เลขบัตรประจำตัวราชการ, ชื่อ, นามสกุล, วันเกิด, เบอร์โทร, อีเมล, ที่อยู่แบบแยกฟิลด์, รูป
- **ตำแหน่งบริหาร** (`admin_positions`): director (ผอ.) **มีได้ 1 คนที่ active** (บังคับด้วย partial unique index), deputy_director (รอง ผอ.) มีได้หลายคน
- **วิทยฐานะ** (`academic_standings`): เก็บเป็นประวัติหลายรายการ (standing, effective_date, is_current) — is_current ได้แค่ 1 ต่อคน
- **ผลงานครู** (`personnel_works`): ผูกเทอม, แนบไฟล์ได้หลายไฟล์ผ่าน `personnel_work_files` (image/document/certificate)

### 4.4 นักเรียน + ผู้ปกครอง (กลุ่มงานวิชาการเป็นผู้เพิ่ม)
- **นักเรียน** (`students`): เลขบัตรประชาชน, รหัสนักเรียน, ชื่อ, นามสกุล, วันเกิด, เบอร์โทร, ที่อยู่แบบแยกฟิลด์, รูป
- **ผู้ปกครอง** (`guardians`): เลขบัตรประชาชน, ชื่อ, นามสกุล, วันเกิด, เบอร์โทร, ที่อยู่แบบแยกฟิลด์
- นักเรียนมีผู้ปกครองได้**หลายคน** (M:N ผ่าน `student_guardians`) โดยตารางเชื่อมเก็บ relationship (father/mother/other) และ is_primary (ผู้ปกครองหลัก)
- ผู้ปกครอง 1 คนผูกกับนักเรียนหลายคนได้ (กรณีพี่น้อง)

### 4.5 ที่อยู่แบบแยกฟิลด์ (มาตรฐานไทย)
ทุกที่อยู่ (personnel/students/guardians) แยกเป็น: บ้านเลขที่ (house_no), หมู่ (moo), ถนน (road), ตำบล/แขวง (subdistrict), อำเภอ/เขต (district), จังหวัด (province), รหัสไปรษณีย์ (postal_code)

### 4.6 ความสัมพันธ์สำคัญ
- ครู 1 คน สอนได้หลายวิชา (M:N `teaching_assignments`)
- ห้องที่ปรึกษา 1 ห้อง มีครูได้หลายคน (M:N `class_advisors`)
- นักเรียน 1 คน อยู่ห้องที่ปรึกษาได้ห้องเดียวต่อเทอม (`student_enrollments` UNIQUE(student_id, semester_id))

### 4.7 การเช็คชื่อมี 2 แบบ
- **เช็คชื่อเข้าเรียน** (`attendances`) — รายวัน ครูที่ปรึกษาเช็ค 1 ครั้ง/วัน
- **เช็คชื่อรายวิชา** (`subject_attendances`) — รายคาบ ครูประจำวิชาเช็ค หลายครั้ง/วัน
- สถานะ: present, absent, late, sick_leave, personal_leave

### 4.8 คะแนนความประพฤติเก็บเป็นประวัติ
`behavior_records` เก็บการหัก/เพิ่มทีละรายการ (ใคร เมื่อไหร่ เหตุผล กี่คะแนน) คะแนนปัจจุบัน = ตั้งต้น + SUM(points)

### 4.9 ระบบสิทธิ์ 3 ชั้น
1. School Admin — ทุกกลุ่มงาน + ตั้งค่าระบบ
2. Group Admin — เฉพาะกลุ่มงานที่ is_group_admin
3. Group Member — ตามกลุ่มงานที่สังกัด
- สิทธิ์ครู (เช็คชื่อ/คะแนนวิชาตัวเอง, ดูนักเรียนที่ปรึกษา) มาจาก role ไม่ใช่กลุ่มงาน
- หน้าตั้งค่าตารางสอน เข้าได้เฉพาะกลุ่มวิชาการ + admin

---

## 5. ข้อมูลส่วนบุคคล / PDPA (บังคับ)

schema มีข้อมูลอ่อนไหวมาก (เลขบัตรประชาชน 13 หลักของครู/นักเรียน/ผู้ปกครอง, ข้อมูลเด็ก) ต้องปฏิบัติตามนี้ทุกครั้งที่แตะข้อมูลส่วนบุคคล:

- **เลขบัตรประชาชน** ต้องเข้ารหัสตอนเก็บ (encryption at rest) หรือจำกัดสิทธิ์เข้าถึงเข้มงวด ห้ามส่งกลับ frontend เต็มเลขถ้าไม่จำเป็น (mask เช่น 1-2345-xxxxx-xx-1)
- **Audit log** — บันทึกทุกครั้งที่มีการดู/แก้/ลบข้อมูลส่วนบุคคล (ใคร เมื่อไหร่ ข้อมูลใคร)
- **จำกัดการเข้าถึงตามกลุ่มงาน** — ข้อมูลครู: กลุ่มบุคคล; ข้อมูลนักเรียน/ผู้ปกครอง: กลุ่มวิชาการ (+ admin)
- **ไฟล์แนบ** (รูป/เอกสาร/เกียรติบัตร) เข้าถึงผ่าน signed URL ที่หมดอายุ ไม่ใช่ public URL
- ดู skill sensitive-data-handling

---

## 6. ลำดับการพัฒนา (Phases)

```
Phase 1 — รากฐาน
  กลุ่มบุคคล: personnel, admin_positions, academic_standings, personnel_works(+files)
  กลุ่มวิชาการ: students, guardians, student_guardians
  + schools, academic_years, semesters, users, auth, work_groups
  + classes, class_advisors, student_enrollments

Phase 2 — เช็คชื่อเข้าเรียน + คะแนนความประพฤติ (บริหารทั่วไป)
  + attendances, behavior_records

Phase 3 — ตารางสอน + เช็คชื่อรายวิชา (วิชาการ)
  + subjects, teaching_assignments, timetable_settings,
    period_definitions, timetables, subject_attendances

Phase 4 (ภายหลัง) — งบประมาณและแผน, เกรดละเอียด, assignment
```

ทำทีละเฟส ห้ามข้าม เว้นแต่ผู้ใช้สั่ง

---

## 7. Docker

ทุก service รันใน container ผ่าน docker-compose

```
services:
  frontend  — Next.js (dev: hot reload ผ่าน volume mount)
  backend   — Go (dev: hot reload ด้วย air)
  postgres  — PostgreSQL (มี named volume เก็บข้อมูลถาวร)
  redis     — Redis
```

กฎ Docker:
- ใช้ **multi-stage build** ลดขนาด image (Go binary เล็ก, Next.js standalone output)
- แยก `docker-compose.yml` (dev) กับ `docker-compose.prod.yml`
- Dev ใช้ volume mount + hot reload; Prod ใช้ image ที่ build แล้ว
- มี `healthcheck` ทุก service และ `depends_on` ให้ start ตามลำดับ (postgres/redis ก่อน backend)
- secrets อยู่ใน `.env` (มี `.env.example` ใน repo, ไม่ commit `.env` จริง)
- migration รันผ่าน service/command แยก ไม่ผูกกับการ start backend ปกติ
- เชื่อม service ด้วยชื่อ service ใน network เดียวกัน (เช่น `postgres:5432`, `redis:6379`)

---

## 8. โครงสร้างโปรเจค

```
chumkosoft/
├── CLAUDE.md              ← ไฟล์นี้ (context หลัก)
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
├── backend/
│   ├── CLAUDE.md          ← กฎเฉพาะ Go
│   └── ...
├── frontend/
│   ├── CLAUDE.md          ← กฎเฉพาะ Next.js
│   └── ...
└── .claude/skills/        ← skills เฉพาะโปรเจค
```

อ่าน backend/CLAUDE.md ก่อนแตะโค้ด Go และ frontend/CLAUDE.md ก่อนแตะโค้ด Next.js เสมอ

---

## 9. ภาษาในการสื่อสาร

- โค้ด, ชื่อตัวแปร, comment เทคนิค: ภาษาอังกฤษ
- ข้อความที่ผู้ใช้ปลายทางเห็น: ภาษาไทย
- การสนทนากับผู้พัฒนา: ภาษาไทยได้
