"use client";

import { UserOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";

import { StudentsListPanel } from "@/features/students/StudentsListPanel";
import { PageHeader } from "@/shared/ui/PageHeader";

export default function StudentListPage(): ReactNode {
  return (
    <div className="flex flex-col gap-5">
      <PageHeader icon={<UserOutlined />} title="นักเรียน" subtitle="ข้อมูลนักเรียนและผู้ปกครอง" />
      <StudentsListPanel />
    </div>
  );
}
