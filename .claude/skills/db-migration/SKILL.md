---
name: db-migration
description: เขียน database migration สำหรับ PostgreSQL ในระบบโรงเรียนด้วย golang-migrate ให้ถูก convention ใช้ skill นี้เสมอเมื่อต้องสร้างตารางใหม่ เพิ่ม/แก้ column เปลี่ยน schema หรือเพิ่ม index แม้ผู้ใช้จะพูดสั้น ๆ ว่า "เพิ่มตาราง" หรือ "แก้ฐานข้อมูล" ก็ตาม เพราะ migration ที่ผิด convention (ลืม school_id, ลืม down, ลืม index, ไม่ใช้ UUID/soft delete) สร้างหนี้เทคนิคที่แก้ทีหลังยากมาก
---

# Database Migration

migration ทุกไฟล์ในโปรเจคนี้ต้องทำตาม convention เดียวกัน เพื่อให้ schema สอดคล้องกับ business logic หลัก (multi-tenant, ผูกเทอม, soft delete)

## รูปแบบไฟล์

ใช้ golang-migrate คู่ up/down เสมอ:
```
migrations/
├── 000001_create_schools.up.sql
├── 000001_create_schools.down.sql
├── 000002_create_users.up.sql
└── 000002_create_users.down.sql
```
- เลขลำดับ 6 หลัก เรียงตาม dependency (ตารางที่ถูกอ้างอิงต้องมาก่อน)
- ชื่อสื่อความหมาย: `create_<table>`, `add_<column>_to_<table>`, `add_index_<...>`

## กฎบังคับทุกตารางข้อมูลโรงเรียน

1. **Primary key เป็น UUID**
```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```

2. **มี `school_id`** (FK ไป schools) ในทุกตารางข้อมูลโรงเรียน

3. **ข้อมูลรายเทอมมี `semester_id`** (FK ไป semesters) — ตัดสินจาก: ค่าเปลี่ยนเมื่อขึ้นเทอมใหม่ไหม (ดู skill tenant-term-scoping)

4. **Timestamp + soft delete** สำหรับข้อมูลสำคัญ
```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
deleted_at TIMESTAMPTZ
```

5. **Index** บน `school_id`, `semester_id`, และ FK ที่ query บ่อย

6. **Constraint สะท้อน business rule** เช่น นักเรียน 1 คน 1 ห้องต่อเทอม:
```sql
UNIQUE (student_id, semester_id)
```

## ตัวอย่างเต็ม — ข้อมูลรายเทอม

`000010_create_attendances.up.sql`:
```sql
CREATE TABLE attendances (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    semester_id UUID NOT NULL REFERENCES semesters(id),
    class_id    UUID NOT NULL REFERENCES classes(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    date        DATE NOT NULL,
    status      VARCHAR(20) NOT NULL
                CHECK (status IN ('present','absent','late','sick_leave','personal_leave')),
    note        TEXT,
    checked_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (student_id, date)
);

CREATE INDEX idx_attendances_school   ON attendances (school_id);
CREATE INDEX idx_attendances_semester ON attendances (semester_id);
CREATE INDEX idx_attendances_class_date ON attendances (class_id, date);
```

`000010_create_attendances.down.sql`:
```sql
DROP TABLE IF EXISTS attendances;
```

## กฎเพิ่มเติม

- **down ต้องย้อน up ได้จริง** — สร้างตารางใน up → drop ใน down; เพิ่ม column ใน up → drop column ใน down
- **ห้ามแก้ migration ที่ถูก apply ไปแล้ว** — สร้างไฟล์ใหม่เพื่อแก้ schema แทน
- **enum/สถานะ** ใช้ `CHECK` constraint หรือ lookup table ให้สอดคล้องกับค่าใน backend (สถานะการลาต้องตรงกับ frontend zod schema ด้วย)
- ใช้ `TIMESTAMPTZ` ไม่ใช่ `TIMESTAMP` (มี timezone)
- seed ข้อมูลตั้งต้น (super admin, work_groups 4 กลุ่ม, โรงเรียนแรก) ทำผ่าน migration แยกที่ชื่อชัด เช่น `000099_seed_initial.up.sql`

## Checklist ก่อนเสร็จ

- [ ] มีทั้งไฟล์ up และ down
- [ ] PK เป็น UUID
- [ ] มี school_id (+ semester_id ถ้ารายเทอม)
- [ ] มี created_at/updated_at/deleted_at (ข้อมูลสำคัญ)
- [ ] มี index บน school_id/semester_id/FK
- [ ] constraint สะท้อน business rule ที่เกี่ยวข้อง
- [ ] down ย้อนกลับได้จริง
- [ ] รัน migrate up แล้ว migrate down แล้ว up ใหม่ ผ่านทั้งหมด
