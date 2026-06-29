DROP INDEX IF EXISTS idx_students_status;
ALTER TABLE students DROP COLUMN IF EXISTS status;
