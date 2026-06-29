-- รายวิชา (ถาวร ไม่ผูกเทอม ตาม CLAUDE.md ข้อ 4.2 "วิชาถาวร")
CREATE TABLE subjects (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    subject_code VARCHAR(50) NOT NULL,            -- เช่น ค21101
    name         VARCHAR(255) NOT NULL,           -- เช่น คณิตศาสตร์พื้นฐาน
    credit       NUMERIC(3,1),                    -- หน่วยกิต (ไม่บังคับ)
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (school_id, subject_code)
);

CREATE INDEX idx_subjects_school ON subjects (school_id);
