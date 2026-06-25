CREATE TABLE user_work_groups (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    user_id        UUID NOT NULL REFERENCES users(id),
    work_group_id  UUID NOT NULL REFERENCES work_groups(id),
    is_group_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    UNIQUE (user_id, work_group_id)
);

CREATE INDEX idx_user_work_groups_school     ON user_work_groups (school_id);
CREATE INDEX idx_user_work_groups_user       ON user_work_groups (user_id);
CREATE INDEX idx_user_work_groups_work_group ON user_work_groups (work_group_id);
