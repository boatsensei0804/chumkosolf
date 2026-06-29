-- เพิ่ม role 'kiosk' (บัญชีสำหรับเครื่องสแกนหน้าเข้าเรียนโดยเฉพาะ — least privilege)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('super_admin', 'teacher', 'executive', 'student', 'kiosk'));
