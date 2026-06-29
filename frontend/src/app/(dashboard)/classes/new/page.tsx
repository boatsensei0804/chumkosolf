"use client";

import { ApartmentOutlined } from "@ant-design/icons";
import { App } from "antd";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { ClassForm, type ClassFormValues } from "@/features/classes/ClassForm";
import { useCreateClass } from "@/features/classes/hooks";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

export default function NewClassPage(): ReactNode {
  const router = useRouter();
  const { message } = App.useApp();
  const createMutation = useCreateClass();
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (v: ClassFormValues): void => {
    setErrorMessage("");
    createMutation.mutate(
      { grade_level: v.gradeLevel, room_name: v.roomName },
      {
        onSuccess: () => {
          message.success("เพิ่มห้องเรียนแล้ว");
          router.push("/classes");
        },
        onError: (err) => setErrorMessage(err.message),
      },
    );
  };

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-5">
      <PageHeader icon={<ApartmentOutlined />} title="เพิ่มห้องเรียน" subtitle="สร้างห้องที่ปรึกษาสำหรับเทอมปัจจุบัน" backHref="/classes" backLabel="กลับไปรายการห้องเรียน" />
      <SectionCard title="ข้อมูลห้องเรียน">
        <ClassForm
          defaultValues={{ gradeLevel: "", roomName: "" }}
          onSubmit={handleSubmit}
          isSubmitting={createMutation.isPending}
          errorMessage={errorMessage}
          submitLabel="เพิ่มห้องเรียน"
        />
      </SectionCard>
    </div>
  );
}
