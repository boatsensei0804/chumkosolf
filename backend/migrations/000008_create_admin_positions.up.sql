CREATE TABLE admin_positions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    personnel_id UUID NOT NULL REFERENCES personnel(id),
    position     VARCHAR(20) NOT NULL
                 CHECK (position IN ('director', 'deputy_director')),
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    appointed_at DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_admin_positions_school    ON admin_positions (school_id);
CREATE INDEX idx_admin_positions_personnel ON admin_positions (personnel_id);
-- ผอ. (director) ที่ active ได้คนเดียวต่อโรงเรียน; รอง ผอ. มีได้หลายคน
CREATE UNIQUE INDEX uq_admin_positions_active_director
    ON admin_positions (school_id)
    WHERE position = 'director' AND is_active = TRUE AND deleted_at IS NULL;
