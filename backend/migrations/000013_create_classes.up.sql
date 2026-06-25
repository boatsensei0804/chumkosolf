-- ห้องที่ปรึกษา (รายเทอม)
CREATE TABLE classes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    semester_id UUID NOT NULL REFERENCES semesters(id),
    grade_level VARCHAR(50) NOT NULL,   -- เช่น ม.1, ป.6
    room_name   VARCHAR(50) NOT NULL,   -- เช่น 1/1, 1/2
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (school_id, semester_id, grade_level, room_name)
);

CREATE INDEX idx_classes_school   ON classes (school_id);
CREATE INDEX idx_classes_semester ON classes (semester_id);
