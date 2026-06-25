"use client";

import { App, Spin } from "antd";
import { useParams, useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { PersonnelForm } from "@/features/personnel/PersonnelForm";
import { PersonnelSubResources } from "@/features/personnel/PersonnelSubResources";
import { WorkGroupsSection } from "@/features/personnel/WorkGroupsSection";
import { toUpdateBody, type CreatePersonnelFormValues } from "@/features/personnel/formSchema";
import { usePersonnel, useUpdatePersonnel } from "@/features/personnel/hooks";
import type { PersonnelDetail } from "@/shared/schemas/personnel";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

// แปลงข้อมูลจาก backend → ค่าเริ่มต้นของฟอร์ม (เลขบัตรเว้นว่าง = ไม่เปลี่ยน, ตาม PDPA ไม่ดึงเลขเต็ม)
function toFormValues(d: PersonnelDetail): CreatePersonnelFormValues {
  return {
    username: d.username,
    password: "",
    role: d.role === "executive" ? "executive" : "teacher",
    nationalId: "",
    civilServantId: "",
    prefix: d.prefix,
    firstName: d.first_name,
    lastName: d.last_name,
    birthDate: d.birth_date,
    phone: d.phone,
    email: d.email,
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

export default function EditPersonnelPage(): ReactNode {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();
  const { message } = App.useApp();

  const { data, isLoading, isError, error } = usePersonnel(id);
  const updateMutation = useUpdatePersonnel(id);
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (values: CreatePersonnelFormValues): void => {
    setErrorMessage("");
    updateMutation.mutate(toUpdateBody(values), {
      onSuccess: () => {
        message.success("บันทึกข้อมูลบุคลากรแล้ว");
        router.push("/personnel");
      },
      onError: (err) => setErrorMessage(err.message),
    });
  };

  const fullName = data ? `${data.prefix}${data.first_name} ${data.last_name}`.trim() : "";

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-5">
      <PageHeader
        title="แก้ไขข้อมูลบุคลากร"
        subtitle={fullName || "กำลังโหลด…"}
        backHref="/personnel"
        backLabel="กลับไปรายการบุคลากร"
      />

      <SectionCard title="ข้อมูลบุคลากร">
        {isLoading && (
          <div className="flex justify-center py-10">
            <Spin />
          </div>
        )}
        {isError && (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
            โหลดข้อมูลไม่สำเร็จ: {error?.message}
          </div>
        )}
        {data && (
          <PersonnelForm
            mode="edit"
            defaultValues={toFormValues(data)}
            onSubmit={handleSubmit}
            isSubmitting={updateMutation.isPending}
            errorMessage={errorMessage}
            submitLabel="บันทึกการแก้ไข"
          />
        )}
      </SectionCard>

      {data && <WorkGroupsSection personnelId={id} />}
      {data && <PersonnelSubResources personnelId={id} />}
    </div>
  );
}
