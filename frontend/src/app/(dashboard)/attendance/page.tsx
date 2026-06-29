"use client";

import { TrophyOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";

import { BehaviorTab } from "@/features/attendance/BehaviorTab";
import { PageHeader } from "@/shared/ui/PageHeader";

export default function BehaviorPage(): ReactNode {
  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<TrophyOutlined />}
        title="คะแนนความประพฤติ"
        subtitle="กลุ่มงานบริหารทั่วไป · บันทึกหัก/เพิ่มคะแนนความประพฤติ"
      />
      <div className="rounded-xl border border-slate-200 bg-white p-4">
        <BehaviorTab />
      </div>
    </div>
  );
}
