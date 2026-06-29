-- เช็คชื่อเข้าเรียน (รายวัน) — ครูที่ปรึกษาเช็ค 1 ครั้ง/วัน/นักเรียน (CLAUDE.md ข้อ 4.7)
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
    -- นักเรียน 1 คนมีผลเช็คชื่อได้ 1 รายการ/วัน (upsert ผ่าน conflict target นี้)
    UNIQUE (student_id, date)
);

CREATE INDEX idx_attendances_school     ON attendances (school_id);
CREATE INDEX idx_attendances_semester   ON attendances (semester_id);
CREATE INDEX idx_attendances_class_date ON attendances (class_id, date);
CREATE INDEX idx_attendances_student    ON attendances (student_id);
