CREATE TABLE work_groups (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    code        VARCHAR(30) NOT NULL
                CHECK (code IN ('personnel', 'general_affairs', 'academic', 'budget_plan')),
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (school_id, code)
);

CREATE INDEX idx_work_groups_school ON work_groups (school_id);
