"use client";

import { SettingOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { isSchoolAdmin } from "@/features/navigation/menu";
import { SchoolSettings } from "@/features/school/SchoolSettings";
import { PageHeader } from "@/shared/ui/PageHeader";

export default function SettingsPage(): ReactNode {
  const { user } = useAuth();
  const allowed = user ? isSchoolAdmin(user) : false;

  return (
    <div className="flex flex-col gap-5">
      <PageHeader icon={<SettingOutlined />} title="ตั้งค่าระบบ" subtitle="สำหรับผู้ดูแลระบบของโรงเรียน" />
      {allowed ? (
        <SchoolSettings />
      ) : (
        <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700">
          หน้านี้สำหรับผู้ดูแลระบบของโรงเรียนเท่านั้น
        </div>
      )}
    </div>
  );
}
