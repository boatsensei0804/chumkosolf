-- pgcrypto ให้ gen_random_uuid() (สำรองสำหรับ PG < 13) และ digest() หากต้องใช้
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE schools (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    code        VARCHAR(50)  NOT NULL,
    address     TEXT,
    phone       VARCHAR(20),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (code)
);
