"use client";

import { CheckCircleFilled, UserAddOutlined } from "@ant-design/icons";
import { App, Button } from "antd";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { enrollStudent } from "@/features/classes/api";
import { emptyAddress } from "@/features/personnel/formSchema";
import { GuardiansSection } from "@/features/students/GuardiansSection";
import { StudentForm } from "@/features/students/StudentForm";
import { StudentPhotoCard } from "@/features/students/StudentPhotoCard";
import { toStudentBody, type CreateStudentFormValues } from "@/features/students/formSchema";
import { useCreateStudent } from "@/features/students/hooks";
import { ApiRequestError } from "@/lib/api/client";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

const emptyValues: CreateStudentFormValues = {
  nationalId: "",
  studentCode: "",
  status: "studying",
  classId: "",
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
  // หลังบันทึกนักเรียนสำเร็จ จะได้ id มาเพิ่มผู้ปกครอง + รูปได้เลยในหน้าเดียวกัน
  const [newStudent, setNewStudent] = useState<{ id: string; name: string } | null>(null);

  const handleSubmit = async (values: CreateStudentFormValues): Promise<void> => {
    setErrorMessage("");
    try {
      const newId = await createMutation.mutateAsync(toStudentBody(values));
      // จัดนักเรียนเข้าห้องของเทอมปัจจุบัน (ถ้าเลือกห้อง)
      if (values.classId !== "") {
        await enrollStudent(values.classId, { student_id: newId });
      }
      message.success("บันทึกนักเรียนแล้ว — เพิ่มผู้ปกครองและรูปได้เลย");
      setNewStudent({ id: newId, name: `${values.prefix}${values.firstName} ${values.lastName}`.trim() });
    } catch (err) {
      setErrorMessage(err instanceof ApiRequestError ? err.message : "เกิดข้อผิดพลาด กรุณาลองใหม่");
    }
  };

  // ขั้นที่ 2: บันทึกนักเรียนแล้ว → เพิ่มผู้ปกครอง + รูป (ใช้ id ที่เพิ่งสร้าง)
  if (newStudent) {
    return (
      <div className="flex flex-col gap-5">
        <PageHeader
          icon={<UserAddOutlined />}
          title="เพิ่มผู้ปกครองและรูปนักเรียน"
          subtitle={newStudent.name}
          backHref="/students"
          backLabel="กลับไปรายการนักเรียน"
        />

        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700">
          <CheckCircleFilled className="text-emerald-500" />
          บันทึกข้อมูลนักเรียนแล้ว ขั้นต่อไปเพิ่มผู้ปกครองและรูป (จะข้ามก็ได้)
        </div>

        <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
          <div className="xl:col-span-2">
            <GuardiansSection studentId={newStudent.id} />
          </div>
          <StudentPhotoCard studentId={newStudent.id} />
        </div>

        <div className="flex justify-end gap-2">
          <Button onClick={() => router.push(`/students/${newStudent.id}/edit`)}>แก้ไขข้อมูลนักเรียนต่อ</Button>
          <Button type="primary" onClick={() => router.push("/students")}>
            เสร็จสิ้น
          </Button>
        </div>
      </div>
    );
  }

  // ขั้นที่ 1: กรอกข้อมูลนักเรียน
  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-5">
      <PageHeader icon={<UserAddOutlined />} title="เพิ่มนักเรียน" subtitle="บันทึกข้อมูลนักเรียน" backHref="/students" backLabel="กลับไปรายการนักเรียน" />
      <SectionCard title="ข้อมูลนักเรียน">
        <StudentForm
          mode="create"
          defaultValues={emptyValues}
          onSubmit={handleSubmit}
          isSubmitting={createMutation.isPending}
          errorMessage={errorMessage}
          submitLabel="บันทึกแล้วเพิ่มผู้ปกครอง/รูป"
        />
      </SectionCard>
    </div>
  );
}
