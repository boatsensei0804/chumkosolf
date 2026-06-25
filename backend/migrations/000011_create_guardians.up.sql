CREATE TABLE guardians (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id           UUID NOT NULL REFERENCES schools(id),

    -- ข้อมูลอ่อนไหว (PDPA)
    national_id_encrypted BYTEA NOT NULL,
    national_id_hash      VARCHAR(64) NOT NULL,

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

    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,

    UNIQUE (school_id, national_id_hash)
);

CREATE INDEX idx_guardians_school ON guardians (school_id);
