"use client";

import { CameraOutlined, VideoCameraOutlined } from "@ant-design/icons";
import { Button, Modal, Select } from "antd";
import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";

const CAM_KEY = "studentPhoto.cameraId";

// CameraCaptureModal — ถ่ายรูปจากกล้อง/webcam (รองรับเลือกกล้องภายนอก) แล้วส่งกลับเป็นไฟล์ JPEG
// ใช้เป็นทางเลือกแทนการอัปโหลดไฟล์ (เช่น เก็บรูปนักเรียนสำหรับสแกนหน้า)
export function CameraCaptureModal(props: {
  open: boolean;
  onClose: () => void;
  onCapture: (file: File) => void;
  busy?: boolean;
}): ReactNode {
  const { open, onClose, onCapture, busy } = props;
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const [error, setError] = useState("");
  const [cameras, setCameras] = useState<MediaDeviceInfo[]>([]);
  const [cameraId, setCameraId] = useState<string>("");

  const startCamera = useCallback(async (deviceId?: string): Promise<void> => {
    if (!navigator.mediaDevices?.getUserMedia) {
      setError("เบราว์เซอร์นี้ไม่รองรับกล้อง");
      return;
    }
    try {
      streamRef.current?.getTracks().forEach((t) => t.stop());
      const stream = await navigator.mediaDevices.getUserMedia({
        video: deviceId ? { deviceId: { exact: deviceId } } : { facingMode: "user" },
        audio: false,
      });
      streamRef.current = stream;
      if (videoRef.current) videoRef.current.srcObject = stream;
      setError("");
      const devices = await navigator.mediaDevices.enumerateDevices();
      setCameras(devices.filter((d) => d.kind === "videoinput"));
    } catch {
      setError("เปิดกล้องไม่สำเร็จ — โปรดอนุญาตการใช้กล้อง");
    }
  }, []);

  const stopCamera = useCallback((): void => {
    streamRef.current?.getTracks().forEach((t) => t.stop());
    streamRef.current = null;
  }, []);

  useEffect(() => {
    if (!open) return;
    const saved = (typeof window !== "undefined" && window.localStorage?.getItem(CAM_KEY)) || "";
    setCameraId(saved);
    void startCamera(saved || undefined);
    return () => stopCamera();
  }, [open, startCamera, stopCamera]);

  const onPickCamera = (id: string): void => {
    setCameraId(id);
    window.localStorage?.setItem(CAM_KEY, id);
    void startCamera(id || undefined);
  };

  const capture = (): void => {
    const video = videoRef.current;
    const canvas = canvasRef.current;
    if (!video || !canvas || video.videoWidth === 0) return;
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext("2d")?.drawImage(video, 0, 0);
    canvas.toBlob(
      (blob) => {
        if (!blob) return;
        onCapture(new File([blob], `photo-${Date.now()}.jpg`, { type: "image/jpeg" }));
      },
      "image/jpeg",
      0.9,
    );
  };

  return (
    <Modal title="ถ่ายรูปจากกล้อง" open={open} onCancel={onClose} footer={null} destroyOnHidden width={520}>
      <div className="flex flex-col gap-3">
        {cameras.length > 1 && (
          <div className="flex items-center gap-2">
            <VideoCameraOutlined className="text-slate-400" />
            <Select
              value={cameraId || undefined}
              onChange={onPickCamera}
              placeholder="เลือกกล้อง"
              className="flex-1"
              options={cameras.map((c, i) => ({ value: c.deviceId, label: c.label || `กล้อง ${i + 1}` }))}
            />
          </div>
        )}
        <div className="relative overflow-hidden rounded-xl bg-slate-900">
          {/* eslint-disable-next-line jsx-a11y/media-has-caption -- live webcam ไม่มี caption */}
          <video ref={videoRef} autoPlay playsInline muted className="aspect-video w-full object-cover" />
          <canvas ref={canvasRef} className="hidden" />
          {error && <div className="absolute inset-0 flex items-center justify-center p-4 text-center text-sm text-white">{error}</div>}
        </div>
        <div className="flex justify-end gap-2">
          <Button onClick={onClose}>ปิด</Button>
          <Button type="primary" icon={<CameraOutlined />} loading={busy} disabled={!!error} onClick={capture}>
            ถ่ายรูป
          </Button>
        </div>
        <p className="text-center text-xs text-slate-400">ถ่ายได้หลายรูป · หน้าต่างจะยังเปิดอยู่หลังถ่ายเพื่อถ่ายเพิ่ม</p>
      </div>
    </Modal>
  );
}
