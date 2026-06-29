-- embedding ใบหน้าของนักเรียน (1 รูป = 1 embedding) สำหรับระบบสแกนหน้าเข้าเรียน
-- เก็บเป็น float array (real[]) แล้ว match ด้วย cosine ในชั้น service (Go) — ไม่ต้องใช้ pgvector
-- หมายเหตุ PDPA: embedding ใบหน้าเป็นข้อมูลชีวมาตร (biometric) — เข้าถึงจำกัด + เป็นข้อมูล derived (regenerate ได้)
CREATE TABLE student_face_embeddings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    photo_id    UUID NOT NULL REFERENCES student_photos(id),
    embedding   REAL[] NOT NULL,            -- เวกเตอร์ 512 มิติ (ArcFace)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (photo_id)                       -- 1 embedding ต่อ 1 รูป
);

CREATE INDEX idx_student_face_embeddings_school  ON student_face_embeddings (school_id);
CREATE INDEX idx_student_face_embeddings_student ON student_face_embeddings (student_id);
