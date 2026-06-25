-- นักเรียน 1 คนอยู่ห้องที่ปรึกษาได้ห้องเดียวต่อเทอม
CREATE TABLE student_enrollments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    semester_id UUID NOT NULL REFERENCES semesters(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    class_id    UUID NOT NULL REFERENCES classes(id),
    student_no  INT,   -- เลขที่ในห้อง
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (student_id, semester_id)
);

CREATE INDEX idx_student_enrollments_school   ON student_enrollments (school_id);
CREATE INDEX idx_student_enrollments_semester ON student_enrollments (semester_id);
CREATE INDEX idx_student_enrollments_class    ON student_enrollments (class_id);
CREATE INDEX idx_student_enrollments_student  ON student_enrollments (student_id);
