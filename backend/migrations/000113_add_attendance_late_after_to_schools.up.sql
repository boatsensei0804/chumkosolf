-- เวลาตัด "มา/สาย" ของการเช็คชื่อเข้าเรียน (ตั้งค่าได้ต่อโรงเรียน) — รูปแบบ HH:MM โซนเวลาไทย
ALTER TABLE schools ADD COLUMN attendance_late_after VARCHAR(5) NOT NULL DEFAULT '08:00';
