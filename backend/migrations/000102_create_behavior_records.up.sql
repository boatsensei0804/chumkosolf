-- คะแนนความประพฤติ — เก็บเป็นประวัติหัก/เพิ่มทีละรายการ (CLAUDE.md ข้อ 4.8)
-- คะแนนปัจจุบัน = คะแนนตั้งต้น + SUM(points) ของเทอมนั้น
CREATE TABLE behavior_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    semester_id UUID NOT NULL REFERENCES semesters(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    points      INT NOT NULL CHECK (points <> 0),  -- บวก = เพิ่ม, ลบ = หัก
    reason      TEXT NOT NULL,
    recorded_by UUID NOT NULL REFERENCES users(id),
    occurred_at DATE,                              -- วันที่เกิดเหตุ (ไม่บังคับ)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_behavior_records_school   ON behavior_records (school_id);
CREATE INDEX idx_behavior_records_semester ON behavior_records (semester_id);
CREATE INDEX idx_behavior_records_student  ON behavior_records (student_id);
