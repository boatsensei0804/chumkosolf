---
name: go-api-endpoint
description: สร้าง API endpoint ใหม่ใน backend Go (Fiber) ตาม clean architecture ของระบบโรงเรียน ใช้ skill นี้เสมอเมื่อต้องเพิ่ม endpoint, handler, service, หรือ repository method ใหม่ แม้ผู้ใช้จะพูดแค่ "ทำ API สำหรับ X" หรือ "เพิ่มฟังก์ชันดึง/บันทึก Y" ก็ตาม เพราะ endpoint ที่ไม่ทำตามชั้น handler/service/repository, ไม่ scope tenant, หรือ response format ไม่สม่ำเสมอ จะทำให้ codebase ไม่สอดคล้องและเสี่ยงข้อมูลรั่ว
---

# Go API Endpoint

ทุก endpoint ต้องไหลผ่าน 3 ชั้น: handler → service → repository โดยแต่ละชั้นมีหน้าที่ชัดเจน และต้อง scope ด้วย school_id/semester_id เสมอ (ใช้คู่กับ skill `tenant-term-scoping`)

## หน้าที่ของแต่ละชั้น

- **handler** — parse request, validate รูปแบบ, ดึง scope จาก context, เรียก service, แปลงเป็น response มาตรฐาน ไม่มี business logic
- **service** — business logic, ตรวจสิทธิ์เชิงธุรกิจ, เรียก repository, จัดการ transaction
- **repository** — เข้าถึง DB เท่านั้น, filter school_id/semester_id ทุก query, ไม่มี business logic

## Response format มาตรฐาน

ทุก endpoint ตอบรูปแบบเดียวกัน:
```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Error   *APIError   `json:"error"`
    Meta    *Meta       `json:"meta,omitempty"`
}
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"` // ภาษาไทย ผู้ใช้เข้าใจได้
}
```

## โครงสร้างตัวอย่าง — POST บันทึกเช็คชื่อทั้งห้อง

### handler
```go
func (h *AttendanceHandler) BulkUpsert(c *fiber.Ctx) error {
    ctx := c.UserContext()
    schoolID := tenant.SchoolIDFromContext(ctx)   // จาก context เท่านั้น
    semesterID := tenant.SemesterIDFromContext(ctx)

    var req BulkAttendanceRequest
    if err := c.BodyParser(&req); err != nil {
        return respondError(c, fiber.StatusBadRequest, "INVALID_INPUT", "ข้อมูลไม่ถูกต้อง")
    }

    if err := h.svc.BulkUpsert(ctx, schoolID, semesterID, req); err != nil {
        return respondServiceError(c, err)
    }
    return respondOK(c, fiber.Map{"saved": len(req.Records)})
}
```

### service
```go
func (s *AttendanceService) BulkUpsert(
    ctx context.Context, schoolID, semesterID string, req BulkAttendanceRequest,
) error {
    // ตรวจสิทธิ์: ต้องเป็นครูที่ปรึกษาของห้องนี้ หรืออยู่กลุ่มบริหารทั่วไป
    if err := s.ensureCanCheckClass(ctx, schoolID, req.ClassID); err != nil {
        return err
    }
    // business rule + เรียก repository (ใน transaction)
    return s.repo.BulkUpsert(ctx, schoolID, semesterID, req)
}
```

### repository
ดู skill `tenant-term-scoping` — ทุก query filter school_id + semester_id

## กฎ

1. **scope จาก context เสมอ** ห้ามรับ school_id จาก client (ดู tenant-term-scoping)
2. **validate 2 ชั้น** — รูปแบบที่ handler (validator tags), business rule ที่ service
3. **error มี code + ข้อความไทย** map เป็น HTTP status ที่ถูก (400/401/403/404/409/500)
4. **versioning** ทุก route อยู่ใต้ `/api/v1`
5. **ตั้งสิทธิ์ที่ route** ผ่าน middleware (role/group) ก่อนถึง handler
6. **อัปเดต Swagger** ให้ตรงกับ endpoint จริง

## การทดสอบ (บังคับ)

ทุก endpoint ใหม่ต้องมี test:
- happy path: ส่งข้อมูลถูก → สำเร็จ + response format ถูก
- validation: ส่งข้อมูลผิด → 400
- permission: ผู้ใช้ไม่มีสิทธิ์ → 403
- isolation: ผู้ใช้โรงเรียน A เข้าถึง/แก้ข้อมูลโรงเรียน B ไม่ได้
- รัน `go test ./...`, `golangci-lint run`, `go vet ./...` ผ่านก่อนถือว่าเสร็จ

## Checklist

- [ ] แยกครบ 3 ชั้น (handler/service/repository)
- [ ] scope school_id/semester_id จาก context
- [ ] response format มาตรฐาน
- [ ] error ภาษาไทย + status ถูก
- [ ] route อยู่ใต้ /api/v1 + middleware สิทธิ์
- [ ] มี test ครบ (happy/validation/permission/isolation)
- [ ] Swagger ตรงกับจริง
