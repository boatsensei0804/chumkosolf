"use client";

import { SolutionOutlined } from "@ant-design/icons";
import { Tabs } from "antd";
import type { ReactNode } from "react";

import { AttendanceTab } from "@/features/attendance/AttendanceTab";
import { SubjectAttendanceTab } from "@/features/timetable/SubjectAttendanceTab";
import { PageHeader } from "@/shared/ui/PageHeader";

export default function CheckInPage(): ReactNode {
  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        icon={<SolutionOutlined />}
        title="เช็คชื่อ"
        subtitle="เช็คชื่อเข้าเรียนรายวัน และเช็คชื่อรายวิชา (รายคาบ)"
      />
      <div className="rounded-xl border border-slate-200 bg-white px-4 pb-4">
        <Tabs
          items={[
            { key: "daily", label: "เช็คชื่อเข้าเรียน (รายวัน)", children: <AttendanceTab /> },
            { key: "subject", label: "เช็คชื่อรายวิชา (รายคาบ)", children: <SubjectAttendanceTab /> },
          ]}
        />
      </div>
    </div>
  );
}
