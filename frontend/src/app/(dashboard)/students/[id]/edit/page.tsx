"use client";

import { App, Spin } from "antd";
import { useParams, useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { StudentForm } from "@/features/students/StudentForm";
import { GuardiansSection } from "@/features/students/GuardiansSection";
import { toStudentBody, type CreateStudentFormValues } from "@/features/students/formSchema";
import { useStudent, useUpdateStudent } from "@/features/students/hooks";
import type { StudentDetail } from "@/shared/schemas/student";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

function toFormValues(d: StudentDetail): CreateStudentFormValues {
  return {
    nationalId: "",
    studentCode: d.student_code,
    prefix: d.prefix,
    firstName: d.first_name,
    lastName: d.last_name,
    birthDate: d.birth_date,
    phone: d.phone,
    address: {
      houseNo: d.address.house_no,
      moo: d.address.moo,
      road: d.address.road,
      subdistrict: d.address.subdistrict,
      district: d.address.district,
      province: d.address.province,
      postalCode: d.address.postal_code,
    },
  };
}

export default function EditStudentPage(): ReactNode {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();
  const { message } = App.useApp();

  const { data, isLoading, isError, error } = useStudent(id);
  const updateMutation = useUpdateStudent(id);
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (values: CreateStudentFormValues): void => {
    setErrorMessage("");
    updateMutation.mutate(toStudentBody(values), {
      onSuccess: () => {
        message.success("บันทึกข้อมูลนักเรียนแล้ว");
        router.push("/students");
      },
      onError: (err) => setErrorMessage(err.message),
    });
  };

  const fullName = data ? `${data.prefix}${data.first_name} ${data.last_name}`.trim() : "";

  return (
    <div className="flex flex-col gap-5">
      <PageHeader title="แก้ไขข้อมูลนักเรียน" subtitle={fullName || "กำลังโหลด…"} backHref="/students" backLabel="กลับไปรายการนักเรียน" />

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <div className="xl:col-span-2">
          <SectionCard title="ข้อมูลนักเรียน">
            {isLoading && (
              <div className="flex justify-center py-10">
                <Spin />
              </div>
            )}
            {isError && <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">โหลดข้อมูลไม่สำเร็จ: {error?.message}</div>}
            {data && (
              <StudentForm
                mode="edit"
                defaultValues={toFormValues(data)}
                onSubmit={handleSubmit}
                isSubmitting={updateMutation.isPending}
                errorMessage={errorMessage}
                submitLabel="บันทึกการแก้ไข"
              />
            )}
          </SectionCard>
        </div>

        {data && (
          <div className="flex flex-col gap-5">
            <GuardiansSection studentId={id} />
          </div>
        )}
      </div>
    </div>
  );
}
