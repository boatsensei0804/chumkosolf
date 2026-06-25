-- Audit log การเข้าถึงข้อมูลส่วนบุคคล (PDPA) — append-only, ไม่เก็บค่าจริงของข้อมูลอ่อนไหว
CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID NOT NULL REFERENCES schools(id),
    actor_user_id UUID NOT NULL REFERENCES users(id),
    action        VARCHAR(20) NOT NULL
                  CHECK (action IN ('view', 'create', 'update', 'delete', 'export')),
    target_type   VARCHAR(50) NOT NULL,   -- personnel, student, guardian, ...
    target_id     UUID,
    detail        JSONB,                  -- field ที่แตะ (ไม่เก็บค่าจริงของข้อมูลอ่อนไหว)
    ip_address    INET,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_school ON audit_logs (school_id);
CREATE INDEX idx_audit_logs_actor  ON audit_logs (actor_user_id);
CREATE INDEX idx_audit_logs_target ON audit_logs (target_type, target_id);
CREATE INDEX idx_audit_logs_created ON audit_logs (created_at);
