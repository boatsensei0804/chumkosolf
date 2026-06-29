-- นิยามคาบเรียนรายเทอม (เวลาเริ่ม-จบของแต่ละคาบ, รองรับคาบพัก)
CREATE TABLE period_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    semester_id UUID NOT NULL REFERENCES semesters(id),
    period_no   INT NOT NULL CHECK (period_no BETWEEN 1 AND 20),
    label       VARCHAR(50),            -- เช่น "คาบ 1", "พักเที่ยง"
    start_time  TIME,
    end_time    TIME,
    is_break    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (school_id, semester_id, period_no)
);

CREATE INDEX idx_period_definitions_school   ON period_definitions (school_id);
CREATE INDEX idx_period_definitions_semester ON period_definitions (semester_id);
