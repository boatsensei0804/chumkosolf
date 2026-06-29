-- รูปนักเรียน (หลายรูปต่อคน เพื่อความแม่นยำของระบบสแกนหน้าเข้าเรียน)
-- เข้าถึงผ่าน signed URL เท่านั้น (ไม่ใช่ public bucket) ตาม PDPA — ข้อมูลเด็ก
-- รูปโปรไฟล์ = แถวที่ is_primary (มีได้ 1 รูปต่อนักเรียน) และ sync ไปที่ students.photo_path
CREATE TABLE student_photos (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    student_id   UUID NOT NULL REFERENCES students(id),
    storage_path TEXT NOT NULL,            -- path ใน storage (ไม่ใช่ public URL)
    content_type VARCHAR(100),
    size_bytes   BIGINT,
    is_primary   BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_student_photos_school  ON student_photos (school_id);
CREATE INDEX idx_student_photos_student ON student_photos (student_id);

-- รูปโปรไฟล์ได้แค่ 1 รูปต่อนักเรียน (เฉพาะที่ยังไม่ถูกลบ)
CREATE UNIQUE INDEX uq_student_photos_primary
    ON student_photos (student_id) WHERE is_primary AND deleted_at IS NULL;
