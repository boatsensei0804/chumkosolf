-- นักเรียน M:N ผู้ปกครอง; เก็บ relationship + is_primary (ผู้ปกครองหลัก)
CREATE TABLE student_guardians (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    student_id   UUID NOT NULL REFERENCES students(id),
    guardian_id  UUID NOT NULL REFERENCES guardians(id),
    relationship VARCHAR(20) NOT NULL
                 CHECK (relationship IN ('father', 'mother', 'other')),
    is_primary   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (student_id, guardian_id)
);

CREATE INDEX idx_student_guardians_school   ON student_guardians (school_id);
CREATE INDEX idx_student_guardians_student  ON student_guardians (student_id);
CREATE INDEX idx_student_guardians_guardian ON student_guardians (guardian_id);
-- ผู้ปกครองหลักมีได้คนเดียวต่อนักเรียน
CREATE UNIQUE INDEX uq_student_guardians_primary
    ON student_guardians (student_id)
    WHERE is_primary = TRUE AND deleted_at IS NULL;
