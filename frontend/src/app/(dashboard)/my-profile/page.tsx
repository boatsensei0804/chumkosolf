"use client";

import { IdcardOutlined } from "@ant-design/icons";
import { App, Spin, Tag } from "antd";
import type { ReactNode } from "react";
import { useState } from "react";

import { MyProfileForm } from "@/features/me/MyProfileForm";
import { toMyProfileFormValues, toUpdateMyProfileBody, type MyProfileFormValues } from "@/features/me/formSchema";
import { useMyProfile, useUpdateMyProfile } from "@/features/me/hooks";
import { personnelRoleLabel, type PersonnelRole } from "@/shared/schemas/personnel";
import { PageHeader } from "@/shared/ui/PageHeader";
import { SectionCard } from "@/shared/ui/SectionCard";

function roleText(role: string): string {
  if (role === "teacher" || role === "executive") {
    return personnelRoleLabel[role as PersonnelRole];
  }
  return role;
}

export default function MyProfilePage(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading, isError, error } = useMyProfile();
  const updateMutation = useUpdateMyProfile();
  const [errorMessage, setErrorMessage] = useState("");

  const handleSubmit = (values: MyProfileFormValues): void => {
    setErrorMessage("");
    updateMutation.mutate(toUpdateMyProfileBody(values), {
      onSuccess: () => message.success("บันทึกข้อมูลของคุณแล้ว"),
      onError: (err) => setErrorMessage(err.message),
    });
  };

  const fullName = data ? `${data.prefix}${data.first_name} ${data.last_name}`.trim() : "";

  return (
    <div className="flex flex-col gap-5">
      <PageHeader icon={<IdcardOutlined />} title="ข้อมูลของฉัน" subtitle={fullName || "กำลังโหลด…"} />

      {isLoading && (
        <div className="flex justify-center py-10">
          <Spin />
        </div>
      )}

      {isError && (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-600">
          {error?.code === "PROFILE_NOT_FOUND"
            ? "ไม่พบข้อมูลบุคลากรของคุณในระบบ — โปรดติดต่อกลุ่มงานบุคคล"
            : `โหลดข้อมูลไม่สำเร็จ: ${error?.message}`}
        </div>
      )}

      {data && (
        <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
          {/* คอลัมน์หลัก: ฟอร์มแก้ไขข้อมูลส่วนตัว/ที่อยู่ */}
          <div className="xl:col-span-2">
            <SectionCard title="แก้ไขข้อมูลของฉัน">
              <MyProfileForm
                defaultValues={toMyProfileFormValues(data)}
                onSubmit={handleSubmit}
                isSubmitting={updateMutation.isPending}
                errorMessage={errorMessage}
              />
            </SectionCard>
          </div>

          {/* rail ขวา: ข้อมูลบัญชี (read-only — แก้ที่กลุ่มบุคคลเท่านั้น) */}
          <div className="flex flex-col gap-5">
            <SectionCard icon={<IdcardOutlined />} title="ข้อมูลบัญชี" accent="slate">
              <dl className="flex flex-col gap-3 text-sm">
                <div className="flex items-center justify-between">
                  <dt className="text-slate-500">ชื่อผู้ใช้</dt>
                  <dd className="font-medium text-slate-800">@{data.username}</dd>
                </div>
                <div className="flex items-center justify-between">
                  <dt className="text-slate-500">ตำแหน่ง</dt>
                  <dd>
                    <Tag color={data.role === "executive" ? "purple" : "blue"} bordered={false}>
                      {roleText(data.role)}
                    </Tag>
                  </dd>
                </div>
                <div className="flex items-center justify-between">
                  <dt className="text-slate-500">เลขบัตรประชาชน</dt>
                  <dd className="num text-slate-600">{data.national_id_masked || "—"}</dd>
                </div>
                <div className="flex items-center justify-between">
                  <dt className="text-slate-500">เลขบัตรราชการ</dt>
                  <dd className="num text-slate-600">{data.civil_servant_id_masked || "—"}</dd>
                </div>
              </dl>
              <p className="mt-4 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-400">
                ข้อมูลบัญชีและเลขบัตรแก้ไขได้ที่กลุ่มงานบุคคลเท่านั้น
              </p>
            </SectionCard>
          </div>
        </div>
      )}
    </div>
  );
}
