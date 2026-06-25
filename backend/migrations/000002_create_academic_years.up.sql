CREATE TABLE academic_years (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    year        INT  NOT NULL,              -- ปีการศึกษา พ.ศ. เช่น 2568
    is_current  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (school_id, year)
);

CREATE INDEX idx_academic_years_school ON academic_years (school_id);
-- ปีการศึกษาปัจจุบันมีได้ปีเดียวต่อโรงเรียน
CREATE UNIQUE INDEX uq_academic_years_current
    ON academic_years (school_id)
    WHERE is_current = TRUE AND deleted_at IS NULL;
