"use client";

import { CameraOutlined, DeleteOutlined, StarFilled, StarOutlined, UploadOutlined } from "@ant-design/icons";
import { App, Button, Empty, Spin, Tag, Upload } from "antd";
import { useState, type ReactNode } from "react";

import type { StudentPhoto } from "@/shared/schemas/student";
import { CameraCaptureModal } from "@/shared/ui/CameraCaptureModal";
import { SectionCard } from "@/shared/ui/SectionCard";

import {
  useDeleteStudentPhoto,
  useSetStudentPhotoPrimary,
  useStudentPhotos,
  useUploadStudentPhoto,
} from "./photoHooks";

const ACCEPT = "image/jpeg,image/png,image/webp";
const MAX_BYTES = 5 * 1024 * 1024;
const MAX_PHOTOS = 10;

function PhotoTile({
  photo,
  onSetPrimary,
  onDelete,
  busy,
}: {
  photo: StudentPhoto;
  onSetPrimary: (id: string) => void;
  onDelete: (id: string) => void;
  busy: boolean;
}): ReactNode {
  return (
    <div
      className={`group relative overflow-hidden rounded-xl border bg-slate-50 ${
        photo.is_primary ? "border-brand ring-2 ring-brand/30" : "border-slate-200"
      }`}
    >
      <div className="aspect-square w-full">
        {/* eslint-disable-next-line @next/next/no-img-element -- signed URL จาก storage (ไม่ผ่าน next/image) */}
        <img src={photo.url} alt="รูปนักเรียน" className="h-full w-full object-cover" />
      </div>

      {photo.is_primary && (
        <Tag color="blue" bordered={false} className="absolute left-1.5 top-1.5 m-0">
          <StarFilled className="mr-0.5" />
          โปรไฟล์
        </Tag>
      )}

      <div className="absolute inset-x-0 bottom-0 flex justify-end gap-1 bg-gradient-to-t from-black/55 to-transparent p-1.5 opacity-0 transition-opacity group-hover:opacity-100">
        {!photo.is_primary && (
          <Button
            size="small"
            type="primary"
            icon={<StarOutlined />}
            loading={busy}
            title="ตั้งเป็นรูปโปรไฟล์"
            aria-label="ตั้งเป็นรูปโปรไฟล์"
            onClick={() => onSetPrimary(photo.id)}
          />
        )}
        <Button
          size="small"
          danger
          icon={<DeleteOutlined />}
          title="ลบรูป"
          aria-label="ลบรูป"
          onClick={() => onDelete(photo.id)}
        />
      </div>
    </div>
  );
}

export function StudentPhotoCard({ studentId }: { studentId: string }): ReactNode {
  const { message, modal } = App.useApp();
  const { data, isLoading } = useStudentPhotos(studentId);
  const uploadMutation = useUploadStudentPhoto(studentId);
  const primaryMutation = useSetStudentPhotoPrimary(studentId);
  const deleteMutation = useDeleteStudentPhoto(studentId);

  const [cameraOpen, setCameraOpen] = useState(false);

  const photos = data ?? [];
  const full = photos.length >= MAX_PHOTOS;

  const uploadFile = (file: File): void => {
    uploadMutation.mutate(file, {
      onSuccess: () => message.success("เพิ่มรูปแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  const beforeUpload = (file: File): boolean => {
    if (!ACCEPT.split(",").includes(file.type)) {
      message.error("รองรับเฉพาะรูป JPG, PNG หรือ WEBP");
      return false;
    }
    if (file.size > MAX_BYTES) {
      message.error("รูปต้องมีขนาดไม่เกิน 5 MB");
      return false;
    }
    uploadFile(file);
    return false; // กัน antd อัปโหลดเอง
  };

  const handleSetPrimary = (id: string): void => {
    primaryMutation.mutate(id, {
      onSuccess: () => message.success("ตั้งเป็นรูปโปรไฟล์แล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  // ใช้ modal.confirm (imperative) แทน Popconfirm ต่อรูป — เลี่ยง warning ของ antd ตอน tile unmount หลังลบ
  const handleDelete = (id: string): void => {
    void modal.confirm({
      title: "ลบรูปนี้?",
      okText: "ลบ",
      cancelText: "ยกเลิก",
      okButtonProps: { danger: true },
      onOk: () =>
        new Promise<void>((resolve) => {
          deleteMutation.mutate(id, {
            onSuccess: () => {
              message.success("ลบรูปแล้ว");
              resolve();
            },
            onError: (err) => {
              message.error(err.message);
              resolve();
            },
          });
        }),
    });
  };

  return (
    <SectionCard
      icon={<CameraOutlined />}
      title="รูปนักเรียน"
      description="เก็บได้หลายรูปเพื่อความแม่นยำของระบบสแกนหน้า · เลือก 1 รูปเป็นรูปโปรไฟล์"
      accent="blue"
    >
      {isLoading ? (
        <div className="flex justify-center py-8">
          <Spin />
        </div>
      ) : photos.length === 0 ? (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="ยังไม่มีรูปนักเรียน" />
      ) : (
        <div className="grid grid-cols-3 gap-2">
          {photos.map((p) => (
            <PhotoTile
              key={p.id}
              photo={p}
              onSetPrimary={handleSetPrimary}
              onDelete={handleDelete}
              busy={primaryMutation.isPending}
            />
          ))}
        </div>
      )}

      <div className="mt-3 flex flex-col gap-1.5">
        <div className="flex gap-1.5">
          <Upload accept={ACCEPT} showUploadList={false} beforeUpload={beforeUpload} multiple maxCount={MAX_PHOTOS} className="flex-1 [&_.ant-upload]:block">
            <Button type="primary" icon={<UploadOutlined />} loading={uploadMutation.isPending} disabled={full} block>
              อัปโหลดรูป
            </Button>
          </Upload>
          <Button icon={<CameraOutlined />} disabled={full} onClick={() => setCameraOpen(true)}>
            ถ่ายรูป
          </Button>
        </div>
        <p className="text-center text-xs text-slate-400">
          {full ? "ครบ 10 รูปแล้ว" : `${photos.length}/${MAX_PHOTOS} รูป · JPG/PNG/WEBP ไม่เกิน 5 MB`}
        </p>
      </div>

      <CameraCaptureModal
        open={cameraOpen}
        onClose={() => setCameraOpen(false)}
        onCapture={uploadFile}
        busy={uploadMutation.isPending}
      />
    </SectionCard>
  );
}
