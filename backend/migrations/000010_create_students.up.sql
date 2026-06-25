CREATE TABLE students (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id           UUID NOT NULL REFERENCES schools(id),
    user_id             UUID REFERENCES users(id),  -- นักเรียน login ได้ (อาจตั้งภายหลัง)

    -- ข้อมูลอ่อนไหว (PDPA) — ข้อมูลเด็ก ระวังเป็นพิเศษ
    national_id_encrypted BYTEA NOT NULL,
    national_id_hash      VARCHAR(64) NOT NULL,

    student_code VARCHAR(50) NOT NULL,   -- รหัสนักเรียน
    prefix       VARCHAR(50),
    first_name   VARCHAR(150) NOT NULL,
    last_name    VARCHAR(150) NOT NULL,
    birth_date   DATE,
    phone        VARCHAR(20),

    house_no     VARCHAR(50),
    moo          VARCHAR(50),
    road         VARCHAR(150),
    subdistrict  VARCHAR(150),
    district     VARCHAR(150),
    province     VARCHAR(150),
    postal_code  VARCHAR(10),

    photo_path   TEXT,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,

    UNIQUE (school_id, student_code),
    UNIQUE (school_id, national_id_hash)
);

CREATE INDEX idx_students_school ON students (school_id);
CREATE INDEX idx_students_user   ON students (user_id);
