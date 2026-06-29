"use client";

import { CheckCircleFilled, ClockCircleOutlined, CloseCircleFilled, InfoCircleFilled, KeyOutlined, ReloadOutlined, ScanOutlined, VideoCameraOutlined } from "@ant-design/icons";
import { App, Button, Select, Switch, Tabs, Tag } from "antd";
import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { useRecognizeFace, useReindexFace } from "@/features/face/hooks";
import { KioskAccountsPanel } from "@/features/kiosk-accounts/KioskAccountsPanel";
import { isSchoolAdmin } from "@/features/navigation/menu";
import type { RecognizeResult } from "@/shared/schemas/face";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

type ScanLog = RecognizeResult & { at: string; key: number };

const CAM_KEY = "kiosk.cameraId";

// นาฬิกาเรียลไทม์สำหรับหน้าสแกน (เริ่มหลัง mount เพื่อเลี่ยง hydration mismatch)
function KioskClock(): ReactNode {
  const [now, setNow] = useState<Date | null>(null);
  useEffect(() => {
    setNow(new Date());
    const id = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(id);
  }, []);
  if (!now) return null;
  const time = now.toLocaleTimeString("th-TH", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
  const date = now.toLocaleDateString("th-TH", { weekday: "long", day: "numeric", month: "long", year: "numeric" });
  return (
    <div className="flex items-center justify-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-2.5">
      <ClockCircleOutlined className="text-xl text-sky-500" />
      <span className="num text-2xl font-bold leading-none tracking-tight text-slate-800">{time}</span>
      <span className="text-sm text-slate-500">{date}</span>
    </div>
  );
}

export default function KioskPage(): ReactNode {
  const { message } = App.useApp();
  const { user } = useAuth();
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const [camError, setCamError] = useState("");
  const [cameras, setCameras] = useState<MediaDeviceInfo[]>([]);
  const [cameraId, setCameraId] = useState<string>("");
  const [auto, setAuto] = useState(false);
  const [logs, setLogs] = useState<ScanLog[]>([]);
  const [activeTab, setActiveTab] = useState<"scan" | "accounts">("scan");

  const recognize = useRecognizeFace();
  const reindex = useReindexFace();

  const admin = user ? isSchoolAdmin(user) : false;
  const isKioskRole = user?.role === "kiosk";
  const canReindex = !isKioskRole; // กลุ่มวิชาการ/แอดมิน (บัญชี kiosk รีอินเด็กซ์ไม่ได้)

  // เปิดกล้องตาม deviceId ที่เลือก (หรือกล้อง default)
  const startCamera = useCallback(async (deviceId?: string): Promise<void> => {
    if (!navigator.mediaDevices?.getUserMedia) {
      setCamError("เบราว์เซอร์นี้ไม่รองรับกล้อง");
      return;
    }
    try {
      streamRef.current?.getTracks().forEach((t) => t.stop());
      const constraints: MediaStreamConstraints = {
        video: deviceId ? { deviceId: { exact: deviceId } } : { facingMode: "user" },
        audio: false,
      };
      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      streamRef.current = stream;
      if (videoRef.current) videoRef.current.srcObject = stream;
      setCamError("");
      // หลังได้สิทธิ์ จึงจะเห็นชื่อกล้องครบ
      const devices = await navigator.mediaDevices.enumerateDevices();
      setCameras(devices.filter((d) => d.kind === "videoinput"));
    } catch {
      setCamError("เปิดกล้องไม่สำเร็จ — โปรดอนุญาตการใช้กล้อง");
    }
  }, []);

  // เปิดกล้องเฉพาะตอนอยู่แท็บสแกน (แอดมินที่มาจัดการบัญชีจะไม่ถูกขอกล้อง)
  useEffect(() => {
    if (activeTab !== "scan") return;
    const saved = (typeof window !== "undefined" && window.localStorage?.getItem(CAM_KEY)) || "";
    setCameraId(saved);
    void startCamera(saved || undefined);
    return () => streamRef.current?.getTracks().forEach((t) => t.stop());
  }, [activeTab, startCamera]);

  const onPickCamera = (id: string): void => {
    setCameraId(id);
    window.localStorage?.setItem(CAM_KEY, id);
    void startCamera(id || undefined);
  };

  // สแกนเฟรมเดียวทันที
  const scan = useCallback((): void => {
    const video = videoRef.current;
    const canvas = canvasRef.current;
    if (!video || !canvas || video.videoWidth === 0 || recognize.isPending || !!camError) return;
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext("2d")?.drawImage(video, 0, 0);
    canvas.toBlob(
      (blob) => {
        if (!blob) return;
        recognize.mutate([blob], {
          onSuccess: (res) =>
            setLogs((prev) => [{ ...res, at: new Date().toLocaleTimeString("th-TH"), key: Date.now() }, ...prev].slice(0, 12)),
          onError: (err) => {
            if (err.code !== "NO_FACE_DETECTED") message.error(err.message);
          },
        });
      },
      "image/jpeg",
      0.9,
    );
  }, [recognize, camError, message]);

  // โหมดสแกนอัตโนมัติ (เฉพาะแท็บสแกน)
  useEffect(() => {
    if (!auto || activeTab !== "scan") return;
    const id = setInterval(scan, 3000);
    return () => clearInterval(id);
  }, [auto, activeTab, scan]);

  const handleReindex = (): void => {
    reindex.mutate(undefined, {
      onSuccess: (r) =>
        message.success(`อัปเดตฐานใบหน้าแล้ว: เก็บ ${r.enrolled} รูป${r.skipped > 0 ? ` · ข้าม ${r.skipped} (ไม่พบใบหน้า)` : ""}`),
      onError: (err) => message.error(err.message),
    });
  };

  const busy = recognize.isPending;
  const last = logs[0];

  const scanPanel = (
    <div className="flex flex-col gap-5">
      <KioskClock />
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
      <SectionCard icon={<ScanOutlined />} title="กล้อง" accent="blue">
        {cameras.length > 1 && (
          <div className="mb-3 flex items-center gap-2">
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
          {busy && <div className="absolute inset-x-0 bottom-0 bg-black/55 p-2 text-center text-sm text-white">กำลังสแกน…</div>}
          {camError && (
            <div className="absolute inset-0 flex items-center justify-center p-4 text-center text-sm text-white">{camError}</div>
          )}
        </div>
        <div className="mt-3 flex items-center justify-between gap-3">
          <Button type="primary" size="large" icon={<ScanOutlined />} loading={busy} onClick={scan} disabled={!!camError}>
            สแกน
          </Button>
          <label className="flex items-center gap-2 text-sm text-slate-600">
            สแกนอัตโนมัติ
            <Switch checked={auto} onChange={setAuto} disabled={!!camError} />
          </label>
        </div>

        {last && (
          <div
            className={`mt-3 flex items-center gap-3 rounded-xl border p-3 ${
              last.matched && last.marked
                ? "border-emerald-200 bg-emerald-50"
                : last.matched && last.already_marked
                  ? "border-sky-200 bg-sky-50"
                  : "border-amber-200 bg-amber-50"
            }`}
          >
            {last.matched && last.marked ? (
              <CheckCircleFilled className="text-2xl text-emerald-500" />
            ) : last.matched && last.already_marked ? (
              <InfoCircleFilled className="text-2xl text-sky-500" />
            ) : (
              <CloseCircleFilled className="text-2xl text-amber-500" />
            )}
            <div className="min-w-0">
              {last.matched ? (
                <>
                  <div className="font-semibold text-slate-800">
                    {last.full_name} <span className="num text-xs text-slate-400">{last.student_code}</span>
                  </div>
                  <div className="text-xs">
                    {last.marked ? (
                      <span className="inline-flex flex-wrap items-center gap-1">
                        <Tag color={last.status === "late" ? "warning" : "success"} bordered={false}>
                          {last.status === "late" ? "สาย" : "มาเรียน"} · {last.class_label}
                        </Tag>
                        {last.penalty_applied > 0 && (
                          <Tag color="error" bordered={false}>
                            หัก {last.penalty_applied} คะแนน
                          </Tag>
                        )}
                      </span>
                    ) : last.already_marked ? (
                      <span className="font-medium text-sky-700">
                        สแกนไปแล้ววันนี้ ({last.status === "late" ? "สาย" : "มาเรียน"}) — ไม่บันทึกซ้ำ
                      </span>
                    ) : (
                      <span className="text-amber-600">{last.reason}</span>
                    )}
                  </div>
                </>
              ) : (
                <div className="text-sm text-amber-700">{last.reason || "ไม่พบนักเรียนที่ตรงกับใบหน้า"}</div>
              )}
            </div>
          </div>
        )}
      </SectionCard>

      <SectionCard title="ประวัติการสแกนล่าสุด">
        {logs.length === 0 ? (
          <p className="py-2 text-sm text-slate-400">ยังไม่มีการสแกน</p>
        ) : (
          <ul className="divide-y divide-slate-100">
            {logs.map((l) => (
              <li key={l.key} className="flex items-center justify-between gap-3 py-2.5 text-sm">
                <span className="min-w-0 truncate">
                  {l.matched && l.marked ? (
                    <span className="text-slate-700">
                      {l.full_name}{" "}
                      <Tag color={l.status === "late" ? "warning" : "success"} bordered={false}>
                        {l.status === "late" ? "สาย" : "มา"}
                      </Tag>
                    </span>
                  ) : l.matched && l.already_marked ? (
                    <span className="text-slate-700">
                      {l.full_name}{" "}
                      <Tag color="processing" bordered={false}>
                        สแกนแล้ว
                      </Tag>
                    </span>
                  ) : (
                    <span className="text-amber-600">{l.matched ? l.reason : "ไม่รู้จัก/ตรวจไม่ผ่าน"}</span>
                  )}
                </span>
                <span className="num shrink-0 text-xs text-slate-400">{l.at}</span>
              </li>
            ))}
          </ul>
        )}
      </SectionCard>
      </div>
    </div>
  );

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<ScanOutlined />}
        title="สแกนหน้าเข้าเรียน"
        subtitle="หันหน้าตรงกล้อง ระบบจะบันทึกการมาเรียนอัตโนมัติ"
        actions={
          canReindex ? (
            <Button icon={<ReloadOutlined />} loading={reindex.isPending} onClick={handleReindex}>
              อัปเดตฐานใบหน้า
            </Button>
          ) : undefined
        }
      />

      {admin ? (
        <Tabs
          activeKey={activeTab}
          onChange={(k) => setActiveTab(k as "scan" | "accounts")}
          items={[
            { key: "scan", label: "สแกนหน้า", icon: <ScanOutlined />, children: scanPanel },
            { key: "accounts", label: "บัญชีสแกนหน้า", icon: <KeyOutlined />, children: <KioskAccountsPanel /> },
          ]}
        />
      ) : (
        scanPanel
      )}

      <p className="text-xs text-slate-400">
        หมายเหตุ: ข้อมูลใบหน้าเป็นข้อมูลชีวมาตรที่ใช้เพื่อการเช็คชื่อเท่านั้น และเข้าถึงได้เฉพาะกลุ่มงานที่เกี่ยวข้อง
      </p>
    </div>
  );
}
