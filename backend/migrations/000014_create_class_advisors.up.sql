-- ห้องที่ปรึกษา 1 ห้องมีครูได้หลายคน (M:N, รายเทอม)
CREATE TABLE class_advisors (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    semester_id  UUID NOT NULL REFERENCES semesters(id),
    class_id     UUID NOT NULL REFERENCES classes(id),
    personnel_id UUID NOT NULL REFERENCES personnel(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (class_id, personnel_id)
);

CREATE INDEX idx_class_advisors_school    ON class_advisors (school_id);
CREATE INDEX idx_class_advisors_semester  ON class_advisors (semester_id);
CREATE INDEX idx_class_advisors_class     ON class_advisors (class_id);
CREATE INDEX idx_class_advisors_personnel ON class_advisors (personnel_id);
