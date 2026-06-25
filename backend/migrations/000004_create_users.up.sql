CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    username        VARCHAR(100) NOT NULL,
    password_hash   TEXT NOT NULL,                 -- bcrypt จากฝั่ง Go เท่านั้น
    role            VARCHAR(20) NOT NULL
                    CHECK (role IN ('super_admin', 'teacher', 'executive', 'student')),
    is_school_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (school_id, username)
);

CREATE INDEX idx_users_school ON users (school_id);
