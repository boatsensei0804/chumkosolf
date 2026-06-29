-- มอบหมายการสอน (รายเทอม) — ครู 1 คนสอนได้หลายวิชา/หลายห้อง (M:N ตาม CLAUDE.md ข้อ 4.6)
-- 1 แถว = "ครูคนนี้สอนวิชานี้ให้ห้องนี้ในเทอมนี้"
CREATE TABLE teaching_assignments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    semester_id  UUID NOT NULL REFERENCES semesters(id),
    personnel_id UUID NOT NULL REFERENCES personnel(id),
    subject_id   UUID NOT NULL REFERENCES subjects(id),
    class_id     UUID NOT NULL REFERENCES classes(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    UNIQUE (semester_id, personnel_id, subject_id, class_id)
);

CREATE INDEX idx_teaching_assignments_school    ON teaching_assignments (school_id);
CREATE INDEX idx_teaching_assignments_semester  ON teaching_assignments (semester_id);
CREATE INDEX idx_teaching_assignments_personnel ON teaching_assignments (personnel_id);
CREATE INDEX idx_teaching_assignments_subject   ON teaching_assignments (subject_id);
CREATE INDEX idx_teaching_assignments_class     ON teaching_assignments (class_id);
