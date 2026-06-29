-- ตั้งค่าตารางสอนรายเทอม (ขนาดตาราง: จำนวนวันเรียน/จำนวนคาบต่อวัน) — 1 ชุดต่อเทอม
CREATE TABLE timetable_settings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    semester_id     UUID NOT NULL REFERENCES semesters(id),
    days_per_week   INT NOT NULL DEFAULT 5 CHECK (days_per_week BETWEEN 1 AND 7),
    periods_per_day INT NOT NULL DEFAULT 8 CHECK (periods_per_day BETWEEN 1 AND 20),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (school_id, semester_id)
);

CREATE INDEX idx_timetable_settings_school   ON timetable_settings (school_id);
CREATE INDEX idx_timetable_settings_semester ON timetable_settings (semester_id);
