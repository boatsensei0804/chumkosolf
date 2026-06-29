# face-svc — InsightFace บน CPU (ONNX)
FROM python:3.11-slim

# build-essential/cmake: คอมไพล์ wheel ของ insightface/stringzilla; libgl1/libglib2.0-0: runtime ของ opencv (cv2)
RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential cmake libgl1 libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# ติดตั้ง numpy + cython ก่อน เพื่อให้ build wheel ของ insightface สำเร็จ
RUN pip install --no-cache-dir "numpy<2" cython

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# pre-download โมเดล buffalo_l ตอน build (กันโหลด ~300MB ตอน request แรก)
RUN python -c "import insightface; m=insightface.app.FaceAnalysis(name='buffalo_l', providers=['CPUExecutionProvider']); m.prepare(ctx_id=-1)"

COPY main.py .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
