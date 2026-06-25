-- ครูและผู้บริหารใช้ตารางโปรไฟล์ personnel ร่วมกัน (ผูก users 1:1 สำหรับ login)
CREATE TABLE personnel (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id             UUID NOT NULL REFERENCES schools(id),
    user_id               UUID NOT NULL REFERENCES users(id),

    -- ข้อมูลอ่อนไหว (PDPA): เก็บเข้ารหัส + hash สำหรับ dedup/ค้นหา ไม่เก็บเลขดิบ
    national_id_encrypted      BYTEA NOT NULL,
    national_id_hash           VARCHAR(64) NOT NULL,
    civil_servant_id_encrypted BYTEA,
    civil_servant_id_hash      VARCHAR(64),

    prefix       VARCHAR(50),
    first_name   VARCHAR(150) NOT NULL,
    last_name    VARCHAR(150) NOT NULL,
    birth_date   DATE,
    phone        VARCHAR(20),
    email        VARCHAR(255),

    -- ที่อยู่แบบแยกฟิลด์ (มาตรฐานไทย)
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

    UNIQUE (user_id),
    UNIQUE (school_id, national_id_hash)
);

CREATE INDEX idx_personnel_school ON personnel (school_id);
CREATE INDEX idx_personnel_user   ON personnel (user_id);
