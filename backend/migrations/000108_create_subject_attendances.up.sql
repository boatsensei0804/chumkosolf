-- เช็คชื่อรายวิชา (รายคาบ) — ครูประจำวิชาเช็ค หลายครั้ง/วัน (CLAUDE.md ข้อ 4.7)
CREATE TABLE subject_attendances (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    semester_id  UUID NOT NULL REFERENCES semesters(id),
    timetable_id UUID NOT NULL REFERENCES timetables(id),   -- คาบไหน (ห้อง+วัน+คาบ+วิชา+ครู)
    student_id   UUID NOT NULL REFERENCES students(id),
    date         DATE NOT NULL,
    status       VARCHAR(20) NOT NULL
                 CHECK (status IN ('present','absent','late','sick_leave','personal_leave')),
    note         TEXT,
    checked_by   UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,
    -- นักเรียน 1 คนมีผลเช็คชื่อได้ 1 รายการต่อคาบต่อวัน (upsert ผ่าน conflict target นี้)
    UNIQUE (timetable_id, student_id, date)
);

CREATE INDEX idx_subject_attendances_school        ON subject_attendances (school_id);
CREATE INDEX idx_subject_attendances_semester      ON subject_attendances (semester_id);
CREATE INDEX idx_subject_attendances_timetable_date ON subject_attendances (timetable_id, date);
CREATE INDEX idx_subject_attendances_student       ON subject_attendances (student_id);
