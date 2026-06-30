"""face-svc — บริการคำนวณ embedding ใบหน้า (InsightFace/ArcFace) แบบ stateless

หน้าที่เดียว: รับรูป → ตรวจจับใบหน้า → คืน embedding 512 มิติ (L2-normalized)
การ enroll / จับคู่ / เขียนเช็คชื่อ ทำที่ backend (Go) ทั้งหมด — service นี้ไม่แตะ DB/storage
"""

import cv2
import numpy as np
from fastapi import FastAPI, File, HTTPException, UploadFile

app = FastAPI(title="chumkosoft face-svc")

_model = None


def get_model():
    """โหลดโมเดลแบบ lazy (ครั้งแรกที่เรียก) — buffalo_l = RetinaFace (detect) + ArcFace (embed)"""
    global _model
    if _model is None:
        import insightface

        m = insightface.app.FaceAnalysis(name="buffalo_l", providers=["CPUExecutionProvider"])
        m.prepare(ctx_id=-1, det_size=(640, 640))  # ctx_id=-1 = CPU
        _model = m
    return _model


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/embed")
async def embed(file: UploadFile = File(...)):
    data = await file.read()
    img = cv2.imdecode(np.frombuffer(data, np.uint8), cv2.IMREAD_COLOR)
    if img is None:
        raise HTTPException(status_code=400, detail="invalid image")

    try:
        faces = get_model().get(img)
    except Exception:
        # เฟรมแปลก/เล็กเกินไป → ถือว่าใช้ไม่ได้ (kiosk จะลองเฟรมถัดไป) แทนที่จะ 500
        raise HTTPException(status_code=422, detail="cannot process image")
    if not faces:
        raise HTTPException(status_code=422, detail="no face detected")

    # เลือกใบหน้าที่ใหญ่ที่สุด (ใกล้กล้องสุด) เมื่อมีหลายหน้า
    faces.sort(key=lambda f: (f.bbox[2] - f.bbox[0]) * (f.bbox[3] - f.bbox[1]), reverse=True)
    f = faces[0]
    emb = f.normed_embedding  # 512 มิติ, L2-normalized → ใช้ cosine ได้ตรง
    # yaw proxy จาก 5 keypoints (ตำแหน่งจมูกเทียบกึ่งกลางตา / ระยะระหว่างตา) — ~0 ตรงหน้า, +/- เมื่อหันหน้า
    yaw = 0.0
    try:
        kps = f.kps  # [left_eye, right_eye, nose, left_mouth, right_mouth]
        eye_cx = (float(kps[0][0]) + float(kps[1][0])) / 2.0
        inter = abs(float(kps[1][0]) - float(kps[0][0])) or 1.0
        yaw = (float(kps[2][0]) - eye_cx) / inter
    except Exception:
        yaw = 0.0
    return {"embedding": [float(x) for x in emb], "faces": len(faces), "yaw": yaw}
