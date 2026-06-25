CREATE TABLE semesters (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID NOT NULL REFERENCES schools(id),
    academic_year_id UUID NOT NULL REFERENCES academic_years(id),
    term             INT  NOT NULL CHECK (term IN (1, 2)),  -- ภาคเรียนที่ 1/2
    start_date       DATE,
    end_date         DATE,
    is_current       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ,
    UNIQUE (academic_year_id, term)
);

CREATE INDEX idx_semesters_school        ON semesters (school_id);
CREATE INDEX idx_semesters_academic_year ON semesters (academic_year_id);
-- เทอมปัจจุบันมีได้เทอมเดียวต่อโรงเรียน
CREATE UNIQUE INDEX uq_semesters_current
    ON semesters (school_id)
    WHERE is_current = TRUE AND deleted_at IS NULL;
