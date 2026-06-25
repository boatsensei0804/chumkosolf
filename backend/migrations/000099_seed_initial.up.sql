-- Seed ข้อมูลตั้งต้นสำหรับ dev: โรงเรียนแรก + 4 กลุ่มงาน
-- (super admin user seed ผ่าน CLI `make seed-admin` เพราะต้อง bcrypt hash จริง)
INSERT INTO schools (id, name, code, is_active)
VALUES ('00000000-0000-0000-0000-000000000001', 'โรงเรียนชุมโคพิทยาคม (dev)', 'CHUMKO', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO work_groups (school_id, code, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'personnel',       'กลุ่มงานบุคคล'),
    ('00000000-0000-0000-0000-000000000001', 'general_affairs', 'กลุ่มงานบริหารทั่วไป'),
    ('00000000-0000-0000-0000-000000000001', 'academic',        'กลุ่มงานวิชาการ'),
    ('00000000-0000-0000-0000-000000000001', 'budget_plan',     'กลุ่มงานงบประมาณและแผน')
ON CONFLICT (school_id, code) DO NOTHING;
