-- คะแนนความประพฤติที่หักเมื่อมาสาย (ตั้งค่าได้ต่อโรงเรียน; 0 = ไม่หัก)
ALTER TABLE schools ADD COLUMN attendance_late_penalty INT NOT NULL DEFAULT 5;
