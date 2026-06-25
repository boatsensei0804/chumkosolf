"use client";

import { UserAddOutlined } from "@ant-design/icons";
import { App } from "antd";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { StudentForm } from "@/features/students/StudentForm";
import { toStudentBody, type CreateStudentFormValues } from "@/features/students/formSchema";
import { useCreateStudent } from "@/features/students/hooks";
import { emptyAddress } from "@/features/personnel/formSchema";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

const emptyValues: CreateStudentFormValues = {
  nationalId: "",
  studentCode: "",
  prefix: "",
  firstName: "",
  lastName: "",
  birthDate: "",
  phone: "",
  address: emptyAddress,
};

export default function NewStudentPage(): ReactNode {
  const router = useRouter();
  const { message } = App.useApp();
  const createMutation = useCreateStudent();
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (values: CreateStudentFormValues): void => {
    setErrorMessage("");
    createMutation.mutate(toStudentBody(values), {
      onSuccess: () => {
        message.success("เพิ่มนักเรียนแล้ว");
        router.push("/students");
      },
      onError: (err) => setErrorMessage(err.message),
    });
  };

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-5">
      <PageHeader icon={<UserAddOutlined />} title="เพิ่มนักเรียน" subtitle="บันทึกข้อมูลนักเรียน" backHref="/students" backLabel="กลับไปรายการนักเรียน" />
      <SectionCard title="ข้อมูลนักเรียน">
        <StudentForm mode="create" defaultValues={emptyValues} onSubmit={handleSubmit} isSubmitting={createMutation.isPending} errorMessage={errorMessage} submitLabel="เพิ่มนักเรียน" />
      </SectionCard>
    </div>
  );
}
