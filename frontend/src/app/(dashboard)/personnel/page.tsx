"use client";

import { TeamOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";

import { PersonnelListPanel } from "@/features/personnel/PersonnelListPanel";
import { PageHeader } from "@/shared/ui/PageHeader";

export default function PersonnelListPage(): ReactNode {
  return (
    <div className="flex flex-col gap-5">
      <PageHeader icon={<TeamOutlined />} title="บุคลากร" subtitle="จัดการข้อมูลครูและบุคลากร" />
      <PersonnelListPanel />
    </div>
  );
}
