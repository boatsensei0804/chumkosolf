-- Seed ปีการศึกษา + เทอมปัจจุบันสำหรับ dev (โรงเรียน CHUMKO)
-- ใช้ fixed UUID เพื่อ idempotent; จำเป็นเพราะข้อมูลรายเทอม (ห้องเรียน/จัดห้อง) ต้องมี semester
INSERT INTO academic_years (id, school_id, year, is_current)
VALUES ('00000000-0000-0000-0000-00000000a001', '00000000-0000-0000-0000-000000000001', 2569, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO semesters (id, school_id, academic_year_id, term, start_date, end_date, is_current)
VALUES ('00000000-0000-0000-0000-00000000a002', '00000000-0000-0000-0000-000000000001',
        '00000000-0000-0000-0000-00000000a001', 1, '2026-05-16', '2026-10-10', TRUE)
ON CONFLICT (id) DO NOTHING;
