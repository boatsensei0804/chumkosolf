-- ผลงานครู (รายเทอม)
CREATE TABLE personnel_works (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    semester_id  UUID NOT NULL REFERENCES semesters(id),
    personnel_id UUID NOT NULL REFERENCES personnel(id),
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    work_date    DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_personnel_works_school    ON personnel_works (school_id);
CREATE INDEX idx_personnel_works_semester  ON personnel_works (semester_id);
CREATE INDEX idx_personnel_works_personnel ON personnel_works (personnel_id);
