-- วิทยฐานะ เก็บเป็นประวัติหลายรายการ; is_current ได้แค่ 1 ต่อคน
CREATE TABLE academic_standings (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    personnel_id   UUID NOT NULL REFERENCES personnel(id),
    standing       VARCHAR(100) NOT NULL,   -- เช่น ครูผู้ช่วย, ครู คศ.1, ชำนาญการ
    effective_date DATE,
    is_current     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_academic_standings_school    ON academic_standings (school_id);
CREATE INDEX idx_academic_standings_personnel ON academic_standings (personnel_id);
CREATE UNIQUE INDEX uq_academic_standings_current
    ON academic_standings (personnel_id)
    WHERE is_current = TRUE AND deleted_at IS NULL;
