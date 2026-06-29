-- ตารางสอน (รายเทอม) — 1 ช่อง = ห้อง × วัน × คาบ → มอบหมายการสอน (วิชา+ครู)
CREATE TABLE timetables (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id              UUID NOT NULL REFERENCES schools(id),
    semester_id            UUID NOT NULL REFERENCES semesters(id),
    class_id               UUID NOT NULL REFERENCES classes(id),
    day_of_week            INT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),  -- 1=จันทร์
    period_no              INT NOT NULL CHECK (period_no BETWEEN 1 AND 20),
    teaching_assignment_id UUID NOT NULL REFERENCES teaching_assignments(id),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ,
    -- ห้องหนึ่งมีได้คาบเดียวต่อช่อง (วัน+คาบ) ในเทอม
    UNIQUE (semester_id, class_id, day_of_week, period_no)
);

CREATE INDEX idx_timetables_school      ON timetables (school_id);
CREATE INDEX idx_timetables_semester    ON timetables (semester_id);
CREATE INDEX idx_timetables_class       ON timetables (class_id);
CREATE INDEX idx_timetables_assignment  ON timetables (teaching_assignment_id);
