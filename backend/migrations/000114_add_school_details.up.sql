-- ข้อมูลโรงเรียนแบบละเอียด: ที่อยู่แยกฟิลด์ (มาตรฐานไทย) + ผู้อำนวยการ + อีเมล + เว็บไซต์
ALTER TABLE schools
    ADD COLUMN house_no      VARCHAR(50),
    ADD COLUMN moo           VARCHAR(50),
    ADD COLUMN road          VARCHAR(100),
    ADD COLUMN subdistrict   VARCHAR(100),
    ADD COLUMN district      VARCHAR(100),
    ADD COLUMN province      VARCHAR(100),
    ADD COLUMN postal_code   VARCHAR(10),
    ADD COLUMN director_name VARCHAR(150),
    ADD COLUMN email         VARCHAR(150),
    ADD COLUMN website       VARCHAR(255);
