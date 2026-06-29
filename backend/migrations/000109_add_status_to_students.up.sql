-- สถานะนักเรียน: กำลังศึกษา / ลาออก / แขวนลอย (default = กำลังศึกษา)
ALTER TABLE students
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'studying'
    CHECK (status IN ('studying', 'resigned', 'suspended'));

CREATE INDEX idx_students_status ON students (school_id, status);
