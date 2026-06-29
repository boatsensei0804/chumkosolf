"use client";

import { CalendarOutlined } from "@ant-design/icons";
import { Select, Tag, Tooltip } from "antd";
import { useQueryClient } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";

import { useSemesters } from "@/features/academic/hooks";
import { useCurrentTerm } from "@/features/term/hooks";
import { getSelectedSemesterId, setSelectedSemesterId } from "@/lib/api/termOverride";
import { semesterLabel } from "@/shared/schemas/academic";

// TermSwitcher ให้เลือก "เทอมทำงาน" (override) — ใช้จัดตารางของเทอมอื่นล่วงหน้าได้
// ข้อมูลทุกหน้าที่ผูกเทอมจะ refetch ตามเทอมที่เลือก
export function TermSwitcher(): ReactNode {
  const qc = useQueryClient();
  const { data: current } = useCurrentTerm();
  const { data: semesters } = useSemesters();
  const [selected, setSelected] = useState(getSelectedSemesterId());

  const currentSemId = current?.has_current ? current.semester_id : "";
  const working = selected !== "" ? selected : currentSemId;
  const isOverriding = selected !== "" && selected !== currentSemId;

  const onChange = (id: string): void => {
    const next = id === currentSemId ? "" : id; // เลือกเทอมปัจจุบัน = ยกเลิก override
    setSelectedSemesterId(next);
    setSelected(next);
    // refetch ข้อมูลทุกหน้าให้ตรงเทอมที่เลือก (header เปลี่ยนแล้ว)
    void qc.invalidateQueries();
  };

  const options = (semesters ?? []).map((s) => ({
    value: s.id,
    label: `${semesterLabel(s)}${s.id === currentSemId ? " (ปัจจุบัน)" : ""}`,
  }));

  return (
    <div className="flex items-center gap-2">
      <Tooltip title="เลือกเทอมทำงาน (จัดข้อมูลของเทอมอื่นได้)">
        <Select<string>
          size="small"
          value={working || undefined}
          onChange={onChange}
          options={options}
          placeholder="เลือกเทอม"
          loading={!semesters}
          suffixIcon={<CalendarOutlined />}
          popupMatchSelectWidth={false}
          style={{ minWidth: 190 }}
          status={isOverriding ? "warning" : ""}
        />
      </Tooltip>
      {isOverriding && (
        <Tag color="warning" className="hidden sm:inline">
          ดูเทอมอื่น
        </Tag>
      )}
    </div>
  );
}
