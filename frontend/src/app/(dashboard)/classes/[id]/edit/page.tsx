"use client";

import { App, Spin } from "antd";
import { useParams, useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { AdvisorsSection } from "@/features/classes/AdvisorsSection";
import { ClassForm, type ClassFormValues } from "@/features/classes/ClassForm";
import { EnrollmentsSection } from "@/features/classes/EnrollmentsSection";
import { useClass, useUpdateClass } from "@/features/classes/hooks";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

export default function EditClassPage(): ReactNode {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();
  const { message } = App.useApp();

  const { data, isLoading, isError, error } = useClass(id);
  const updateMutation = useUpdateClass(id);
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (v: ClassFormValues): void => {
    setErrorMessage("");
    updateMutation.mutate(
      { grade_level: v.gradeLevel, room_name: v.roomName },
      {
        onSuccess: () => {
          message.success("บันทึกห้องเรียนแล้ว");
          router.push("/classes");
        },
        onError: (err) => setErrorMessage(err.message),
      },
    );
  };

  const title = data ? `${data.grade_level} ${data.room_name}` : "กำลังโหลด…";

  return (
    <div className="flex flex-col gap-5">
      <PageHeader title="จัดการห้องเรียน" subtitle={title} backHref="/classes" backLabel="กลับไปรายการห้องเรียน" />

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <div className="xl:col-span-2 flex flex-col gap-5">
          <SectionCard title="ข้อมูลห้องเรียน">
            {isLoading && (
              <div className="flex justify-center py-10">
                <Spin />
              </div>
            )}
            {isError && <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดข้อมูลไม่สำเร็จ: {error?.message}</div>}
            {data && (
              <ClassForm
                defaultValues={{ gradeLevel: data.grade_level, roomName: data.room_name }}
                onSubmit={handleSubmit}
                isSubmitting={updateMutation.isPending}
                errorMessage={errorMessage}
                submitLabel="บันทึกการแก้ไข"
              />
            )}
          </SectionCard>
          {data && <EnrollmentsSection classId={id} />}
        </div>

        {data && (
          <div className="flex flex-col gap-5">
            <AdvisorsSection classId={id} />
          </div>
        )}
      </div>
    </div>
  );
}
