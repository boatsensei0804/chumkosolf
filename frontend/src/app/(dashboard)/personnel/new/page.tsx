"use client";

import { UserAddOutlined } from "@ant-design/icons";
import { App } from "antd";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { PersonnelForm } from "@/features/personnel/PersonnelForm";
import { emptyAddress, toCreateBody, type CreatePersonnelFormValues } from "@/features/personnel/formSchema";
import { useCreatePersonnel } from "@/features/personnel/hooks";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

const emptyValues: CreatePersonnelFormValues = {
  username: "",
  password: "",
  role: "teacher",
  nationalId: "",
  civilServantId: "",
  prefix: "",
  firstName: "",
  lastName: "",
  birthDate: "",
  phone: "",
  email: "",
  address: emptyAddress,
};

export default function NewPersonnelPage(): ReactNode {
  const router = useRouter();
  const { message } = App.useApp();
  const createMutation = useCreatePersonnel();
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (values: CreatePersonnelFormValues): void => {
    setErrorMessage("");
    createMutation.mutate(toCreateBody(values), {
      onSuccess: () => {
        message.success("เพิ่มบุคลากรแล้ว");
        router.push("/personnel");
      },
      onError: (err) => setErrorMessage(err.message),
    });
  };

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-5">
      <PageHeader
        icon={<UserAddOutlined />}
        title="เพิ่มบุคลากร"
        subtitle="สร้างบัญชีผู้ใช้พร้อมข้อมูลบุคลากร"
        backHref="/personnel"
        backLabel="กลับไปรายการบุคลากร"
      />
      <SectionCard title="ข้อมูลบุคลากร">
        <PersonnelForm
          mode="create"
          defaultValues={emptyValues}
          onSubmit={handleSubmit}
          isSubmitting={createMutation.isPending}
          errorMessage={errorMessage}
          submitLabel="เพิ่มบุคลากร"
        />
      </SectionCard>
    </div>
  );
}
