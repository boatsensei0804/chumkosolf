---
name: tenant-term-scoping
description: บังคับการ scope ข้อมูลด้วย school_id (multi-tenant) และ semester_id (รายเทอม) ในทุก query และทุก endpoint ของระบบโรงเรียน ใช้ skill นี้เสมอเมื่อเขียนหรือแก้ repository, query SQL, handler, หรือ service ที่อ่าน/เขียนข้อมูลของโรงเรียน หรือเมื่อสร้างตารางใหม่ แม้ผู้ใช้จะไม่ได้พูดถึง tenant หรือ semester ตรง ๆ ก็ตาม เพราะการลืม filter คือช่องโหว่ข้อมูลรั่วข้ามโรงเรียนหรือข้ามเทอมที่ร้ายแรงที่สุดของระบบนี้
---

# Tenant + Term Scoping

ระบบนี้ออกแบบเป็น multi-tenant (เผื่อหลายโรงเรียน) และทุกข้อมูลผูกกับปีการศึกษา/เทอม การลืม filter `school_id` หรือ `semester_id` แม้แต่ query เดียวทำให้ข้อมูลรั่วข้ามโรงเรียนหรือปนข้ามเทอม ซึ่งเป็นบั๊กที่อันตรายและตรวจจับยาก skill นี้กำหนดวิธีบังคับให้ถูกต้องทุกครั้ง

## หลักการ 2 ระดับ

**ระดับ 1 — school_id (ทุกข้อมูลของโรงเรียน)**
ทุกตารางที่เก็บข้อมูลของโรงเรียนมี `school_id` ทุก query ต้อง filter ด้วยค่านี้เสมอ

**ระดับ 2 — semester_id (เฉพาะข้อมูลรายเทอม)**
ข้อมูลที่เปลี่ยนตามเทอมต้องมีและ filter `semester_id` เพิ่ม

ตารางไหนต้องมีอะไร:

| ประเภท | ตัวอย่างตาราง | school_id | semester_id |
|--------|--------------|:---------:|:-----------:|
| คนถาวร | teachers, students, guardians | ✅ | ❌ |
| วิชาถาวร | subjects | ✅ | ❌ |
| โครงสร้างองค์กร | work_groups | ✅ | ❌ |
| การจัดห้องรายเทอม | classes, class_advisors, student_enrollments | ✅ | ✅ |
| การสอนรายเทอม | teaching_assignments, timetables | ✅ | ✅ |
| ตั้งค่าคาบรายเทอม | timetable_settings, period_definitions | ✅ | ✅ |
| เช็คชื่อ | attendances, subject_attendances | ✅ | ✅ |
| คะแนนประพฤติ | behavior_records | ✅ | ✅ |

ถ้าไม่แน่ใจว่าข้อมูลใหม่เป็นถาวรหรือรายเทอม ให้ถาม: "ค่านี้เปลี่ยนเมื่อขึ้นเทอมใหม่ไหม" ถ้าใช่ = รายเทอม

## กฎการเขียนโค้ด

### 1. school_id และ semester_id มาจาก context เท่านั้น — ห้ามรับจาก client
Handler ต้องดึงจาก context ที่ middleware ตั้งไว้ ห้าม bind จาก request body/query ของผู้ใช้ มิฉะนั้นผู้ใช้ปลอม school_id เพื่อเข้าถึงข้อมูลโรงเรียนอื่นได้

**ผิด:**
```go
schoolID := c.Query("school_id") // ผู้ใช้ปลอมได้
```

**ถูก:**
```go
schoolID := tenant.SchoolIDFromContext(c.UserContext())
```

### 2. Repository รับ scope เป็น parameter และ filter เสมอ
ทุก method ใน repository ที่แตะข้อมูลโรงเรียนต้องรับ `schoolID` (และ `semesterID` ถ้ารายเทอม) แล้ว filter ในทุก query — รวมถึง SELECT, UPDATE, DELETE ไม่ใช่แค่ SELECT

**ตัวอย่าง (ข้อมูลรายเทอม):**
```go
func (r *AttendanceRepo) ListByClassAndDate(
    ctx context.Context, schoolID, semesterID, classID string, date time.Time,
) ([]Attendance, error) {
    const q = `
        SELECT id, student_id, status, note
        FROM attendances
        WHERE school_id = $1 AND semester_id = $2
          AND class_id = $3 AND date = $4
          AND deleted_at IS NULL`
    // ...
}
```

**ตัวอย่าง UPDATE — ต้องมี school_id ใน WHERE ด้วย:**
```go
const q = `
    UPDATE students SET full_name = $1
    WHERE id = $2 AND school_id = $3 AND deleted_at IS NULL`
```
ถ้า WHERE ไม่มี `school_id` ผู้ใช้อาจแก้ข้อมูลของโรงเรียนอื่นผ่าน id ที่เดาได้

### 3. INSERT ต้องเซ็ต school_id (+ semester_id) จาก context
```go
const q = `
    INSERT INTO behavior_records (id, school_id, semester_id, student_id, points, reason, recorded_by)
    VALUES ($1, $2, $3, $4, $5, $6, $7)`
```

### 4. Redis cache key ต้องมี school_id (+ semester_id)
ห้ามใช้ key รวมข้ามโรงเรียน มิฉะนั้น cache ปนกัน
```
ถูก:  school:{schoolID}:semester:{semesterID}:class:{classID}:attendance
ผิด:  class:{classID}:attendance
```

## Checklist ก่อนถือว่า query เสร็จ

ถามตัวเองทุกครั้งที่เขียน/แก้ query:
- [ ] WHERE มี `school_id` ไหม (ทุก SELECT/UPDATE/DELETE)
- [ ] ถ้าเป็นข้อมูลรายเทอม WHERE มี `semester_id` ไหม
- [ ] INSERT เซ็ต `school_id` (+ `semester_id`) จาก context ไหม
- [ ] ค่าพวกนี้มาจาก context ไม่ใช่จาก client ใช่ไหม
- [ ] Redis key (ถ้ามี) มี school_id ไหม
- [ ] มี test ยืนยันว่า scope อื่นเข้าถึงข้อมูลนี้ไม่ได้ไหม

## Test ที่ต้องมี

ทุก repository/endpoint ที่แตะข้อมูลโรงเรียน ต้องมี test อย่างน้อย:
1. โรงเรียน A query แล้วไม่เห็นข้อมูลของโรงเรียน B
2. (ถ้ารายเทอม) query เทอมปัจจุบันไม่เห็นข้อมูลเทอมก่อน
3. UPDATE/DELETE ด้วย id ของอีกโรงเรียน ต้องไม่สำเร็จ (affected rows = 0)
