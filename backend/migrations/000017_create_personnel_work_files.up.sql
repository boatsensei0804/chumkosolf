-- ไฟล์แนบของผลงานครู (หลายไฟล์ต่อผลงาน) — เข้าถึงผ่าน signed URL เท่านั้น
CREATE TABLE personnel_work_files (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id          UUID NOT NULL REFERENCES schools(id),
    personnel_work_id  UUID NOT NULL REFERENCES personnel_works(id),
    file_type          VARCHAR(20) NOT NULL
                       CHECK (file_type IN ('image', 'document', 'certificate')),
    storage_path       TEXT NOT NULL,        -- path ใน storage (ไม่ใช่ public URL)
    original_name      VARCHAR(255),
    content_type       VARCHAR(100),
    size_bytes         BIGINT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX idx_personnel_work_files_school ON personnel_work_files (school_id);
CREATE INDEX idx_personnel_work_files_work   ON personnel_work_files (personnel_work_id);
