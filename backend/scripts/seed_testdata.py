#!/usr/bin/env python3
"""สร้างข้อมูลทดสอบครบวงจรผ่าน API (dev) — รัน: python3 seed_testdata.py
สร้าง: ครู kru.demo (มีตารางสอน) + วิชา + ห้อง + นักเรียน + มอบหมายสอน + ตารางสอน
"""
import json
import urllib.request
import urllib.error

BASE = "http://localhost:8080/api/v1"
TOKEN = None


def api(method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE + path, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if TOKEN:
        req.add_header("Authorization", "Bearer " + TOKEN)
    try:
        with urllib.request.urlopen(req) as r:
            return json.loads(r.read())
    except urllib.error.HTTPError as e:
        msg = e.read().decode()
        return json.loads(msg) if msg.strip().startswith("{") else {"success": False, "error": {"message": msg}}


def ok(resp):
    return isinstance(resp, dict) and resp.get("success")


def data_id(resp, label):
    if ok(resp) and resp.get("data", {}).get("id"):
        return resp["data"]["id"]
    print(f"  ! {label}: {json.dumps(resp.get('error') or resp, ensure_ascii=False)}")
    return None


def main():
    global TOKEN
    login = api("POST", "/auth/login", {"school_code": "CHUMKO", "username": "superadmin", "password": "admin1234"})
    if not ok(login):
        print("login ล้มเหลว:", login)
        return
    TOKEN = login["data"]["access_token"]
    print("✓ login superadmin")

    # --- ครู (มี user kru.demo / demo1234) ---
    teacher = api("POST", "/personnel", {
        "username": "kru.demo", "password": "demo1234", "role": "teacher",
        "national_id": "1100000000099", "prefix": "นาย", "first_name": "เดโม", "last_name": "สอนดี",
    })
    teacher_id = data_id(teacher, "teacher")
    if not teacher_id:
        # อาจมีอยู่แล้ว — หา personnel จากรายการ
        lst = api("GET", "/personnel?page=1&page_size=200")
        for p in (lst.get("data") or []):
            if p.get("username") == "kru.demo":
                teacher_id = p["id"]
        if not teacher_id:
            print("ไม่มี teacher_id ใช้งานต่อไม่ได้")
            return
    print("✓ teacher kru.demo:", teacher_id)
    # มอบกลุ่มงานวิชาการให้ครู (เผื่อทดสอบเมนูวิชาการ)
    wgs = api("GET", "/work-groups")
    acad = next((g["id"] for g in (wgs.get("data") or []) if g.get("code") == "academic"), None)
    if acad:
        api("POST", f"/personnel/{teacher_id}/work-groups", {"work_group_id": acad})

    # --- วิชา ---
    subjects = []
    for code, name in [("ค21101", "คณิตศาสตร์พื้นฐาน"), ("ว21101", "วิทยาศาสตร์พื้นฐาน"), ("ท21101", "ภาษาไทย")]:
        r = api("POST", "/subjects", {"subject_code": code, "name": name, "credit": 1.5})
        sid = data_id(r, f"subject {code}")
        if sid:
            subjects.append((sid, code, name))
    print(f"✓ subjects: {len(subjects)}")

    # --- ห้องเรียน (เทอมปัจจุบัน) ---
    classes = []
    for grade, room in [("ม.1", "1/1"), ("ม.1", "1/2")]:
        r = api("POST", "/classes", {"grade_level": grade, "room_name": room})
        cid = data_id(r, f"class {grade} {room}")
        if cid:
            classes.append((cid, grade, room))
    print(f"✓ classes: {len(classes)}")
    if not classes:
        return
    class1 = classes[0][0]

    # --- นักเรียน + จัดเข้าห้อง 1 ---
    names = [("เด็กชาย", "ก้อง", "ใจดี"), ("เด็กหญิง", "ขวัญ", "เรียนเก่ง"), ("เด็กชาย", "คม", "ตั้งใจ"),
             ("เด็กหญิง", "งาม", "ขยัน"), ("เด็กชาย", "จบ", "ครบถ้วน"), ("เด็กหญิง", "ฉวี", "สดใส")]
    enrolled = 0
    for i, (pre, fn, ln) in enumerate(names, start=1):
        nid = f"110000000{i:04d}"
        r = api("POST", "/students", {
            "national_id": nid, "student_code": f"D{i:03d}", "status": "studying",
            "prefix": pre, "first_name": fn, "last_name": ln,
        })
        stid = data_id(r, f"student {fn}")
        if stid:
            er = api("POST", f"/classes/{class1}/students", {"student_id": stid, "student_no": i})
            if ok(er):
                enrolled += 1
    print(f"✓ students enrolled in {classes[0][1]} {classes[0][2]}: {enrolled}")

    # --- มอบหมายการสอน: ครู สอนทั้ง 3 วิชา ให้ห้อง 1 ---
    tas = []
    for sid, code, _ in subjects:
        r = api("POST", "/teaching-assignments", {"personnel_id": teacher_id, "subject_id": sid, "class_id": class1})
        tid = data_id(r, f"assignment {code}")
        if tid:
            tas.append((tid, code))
    print(f"✓ teaching assignments: {len(tas)}")
    if not tas:
        return

    # --- ตั้งค่าคาบ (5 วัน 6 คาบ, คาบ 4 = พักเที่ยง) ---
    periods = []
    times = [("08:30", "09:20"), ("09:20", "10:10"), ("10:10", "11:00"),
             ("11:00", "12:00"), ("13:00", "13:50"), ("13:50", "14:40")]
    for n, (st, en) in enumerate(times, start=1):
        periods.append({"period_no": n, "label": "พักเที่ยง" if n == 4 else f"คาบ {n}",
                        "start_time": st, "end_time": en, "is_break": n == 4})
    cfg = api("PUT", "/timetable/config", {"days_per_week": 5, "periods_per_day": 6, "periods": periods})
    print("✓ timetable config" if ok(cfg) else f"  ! config: {cfg.get('error')}")

    # --- จัดตารางสอนให้ห้อง 1 (วาง 3 วิชาในหลายคาบ) ---
    placements = [(1, 1, 0), (2, 2, 1), (3, 3, 2), (4, 1, 0), (5, 2, 1)]  # (day, period, ta_index)
    placed = 0
    for day, period, idx in placements:
        if idx < len(tas):
            r = api("POST", f"/timetable/classes/{class1}/slots",
                    {"day_of_week": day, "period_no": period, "teaching_assignment_id": tas[idx][0]})
            if ok(r):
                placed += 1
            else:
                print(f"  ! slot d{day}p{period}: {r.get('error')}")
    print(f"✓ timetable slots placed: {placed}")

    print("\n========================================")
    print("เสร็จสิ้น! ทดสอบได้ที่ http://localhost:3000")
    print("  Admin : CHUMKO / superadmin / admin1234")
    print("  ครู   : CHUMKO / kru.demo / demo1234  (มีตารางสอน 5 คาบ → ทดสอบหน้าเช็คชื่อรายวิชา)")
    print("========================================")


if __name__ == "__main__":
    main()
