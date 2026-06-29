"use client";

import { CheckCircleFilled } from "@ant-design/icons";
import { App, Button, DatePicker, InputNumber, Select, Table, Tag, type TableProps } from "antd";
import dayjs from "dayjs";
import { useState, type ReactNode } from "react";

import { SectionCard } from "@/shared/ui/SectionCard";
import {
  semesterLabel,
  type AcademicYear,
  type CreateSemesterBody,
  type Semester,
} from "@/shared/schemas/academic";

import {
  useCreateSemester,
  useCreateYear,
  useSetCurrentSemester,
  useSetCurrentYear,
  useSemesters,
  useYears,
} from "./hooks";

const thisThaiYear = new Date().getFullYear() + 543;

// ===== ฟอร์มเพิ่มปีการศึกษา (presentational) =====
export function YearAddForm(props: { onAdd: (year: number) => void; submitting: boolean }): ReactNode {
  const [year, setYear] = useState<number>(thisThaiYear);
  const [error, setError] = useState("");

  const submit = (): void => {
    if (!year || year < 2400 || year > 2700) {
      setError("กรุณากรอกปีการศึกษา (พ.ศ.) ให้ถูกต้อง");
      return;
    }
    setError("");
    props.onAdd(year);
  };

  return (
    <div className="mt-4 flex flex-col gap-2 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
      <div className="flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ปีการศึกษา (พ.ศ.)</label>
          <InputNumber value={year} onChange={(v) => setYear(v ?? 0)} style={{ width: 120 }} status={error ? "error" : ""} />
        </div>
        <Button type="primary" loading={props.submitting} onClick={submit}>
          เพิ่มปีการศึกษา
        </Button>
      </div>
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  );
}

// ===== ปีการศึกษา =====
function YearsSection(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useYears();
  const createMutation = useCreateYear();
  const setCurrentMutation = useSetCurrentYear();

  const columns: TableProps<AcademicYear>["columns"] = [
    {
      title: "ปีการศึกษา",
      key: "year",
      render: (_, r) => (
        <span className="font-medium text-slate-800">
          <span className="num">{r.year}</span>
          {r.is_current && (
            <Tag color="success" bordered={false} className="ml-2">
              ปัจจุบัน
            </Tag>
          )}
        </span>
      ),
    },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 150,
      render: (_, r) =>
        r.is_current ? (
          <span className="text-xs text-slate-400">
            <CheckCircleFilled className="mr-1 text-emerald-500" />ปีปัจจุบัน
          </span>
        ) : (
          <Button
            size="small"
            loading={setCurrentMutation.isPending}
            onClick={() =>
              setCurrentMutation.mutate(r.id, {
                onSuccess: () => message.success("ตั้งปีการศึกษาปัจจุบันแล้ว"),
                onError: (err) => message.error(err.message),
              })
            }
          >
            ตั้งเป็นปัจจุบัน
          </Button>
        ),
    },
  ];

  return (
    <SectionCard title="ปีการศึกษา" description="กำหนดปีการศึกษาและปีปัจจุบันของโรงเรียน">
      <div className="overflow-hidden rounded-xl border border-slate-200">
        <Table<AcademicYear>
          rowKey="id"
          size="middle"
          columns={columns}
          dataSource={data ?? []}
          loading={isLoading}
          pagination={false}
          locale={{ emptyText: "ยังไม่มีปีการศึกษา" }}
        />
      </div>
      <YearAddForm
        onAdd={(year) =>
          createMutation.mutate(
            { year },
            { onSuccess: () => message.success("เพิ่มปีการศึกษาแล้ว"), onError: (err) => message.error(err.message) },
          )
        }
        submitting={createMutation.isPending}
      />
    </SectionCard>
  );
}

// ===== ภาคเรียน =====
function SemestersSection(): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useSemesters();
  const { data: years } = useYears();
  const createMutation = useCreateSemester();
  const setCurrentMutation = useSetCurrentSemester();

  const [yearId, setYearId] = useState("");
  const [term, setTerm] = useState<number>(1);
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");

  const handleAdd = (): void => {
    if (yearId === "") {
      message.warning("กรุณาเลือกปีการศึกษา");
      return;
    }
    const body: CreateSemesterBody = { academic_year_id: yearId, term, start_date: startDate, end_date: endDate };
    createMutation.mutate(body, {
      onSuccess: () => {
        message.success("เพิ่มภาคเรียนแล้ว");
        setStartDate("");
        setEndDate("");
      },
      onError: (err) => message.error(err.message),
    });
  };

  const columns: TableProps<Semester>["columns"] = [
    {
      title: "ภาคเรียน",
      key: "term",
      render: (_, r) => (
        <span className="font-medium text-slate-800">
          {semesterLabel(r)}
          {r.is_current && (
            <Tag color="success" bordered={false} className="ml-2">
              ปัจจุบัน
            </Tag>
          )}
        </span>
      ),
    },
    {
      title: "ช่วงเวลา",
      key: "dates",
      render: (_, r) => (
        <span className="num text-xs text-slate-400">
          {r.start_date === "" ? "—" : r.start_date} ถึง {r.end_date === "" ? "—" : r.end_date}
        </span>
      ),
    },
    {
      title: "",
      key: "actions",
      align: "right",
      width: 150,
      render: (_, r) =>
        r.is_current ? (
          <span className="text-xs text-slate-400">
            <CheckCircleFilled className="mr-1 text-emerald-500" />เทอมปัจจุบัน
          </span>
        ) : (
          <Button
            size="small"
            loading={setCurrentMutation.isPending}
            onClick={() =>
              setCurrentMutation.mutate(r.id, {
                onSuccess: () => message.success("ตั้งภาคเรียนปัจจุบันแล้ว"),
                onError: (err) => message.error(err.message),
              })
            }
          >
            ตั้งเป็นปัจจุบัน
          </Button>
        ),
    },
  ];

  return (
    <SectionCard title="ภาคเรียน" description="กำหนดภาคเรียน (1/2) ของแต่ละปี และเทอมปัจจุบัน" accent="violet">
      <div className="overflow-hidden rounded-xl border border-slate-200">
        <Table<Semester>
          rowKey="id"
          size="middle"
          columns={columns}
          dataSource={data ?? []}
          loading={isLoading}
          scroll={{ x: 480 }}
          pagination={false}
          locale={{ emptyText: "ยังไม่มีภาคเรียน" }}
        />
      </div>
      <div className="mt-4 flex flex-wrap items-end gap-3 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
        <div>
          <label className="mb-1 block text-xs text-slate-500">ปีการศึกษา</label>
          <Select
            value={yearId || undefined}
            onChange={setYearId}
            placeholder="เลือกปี"
            style={{ width: 130 }}
            options={(years ?? []).map((y) => ({ value: y.id, label: `${y.year}` }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">ภาคเรียน</label>
          <Select<number>
            value={term}
            onChange={setTerm}
            style={{ width: 110 }}
            options={[
              { value: 1, label: "ภาคเรียนที่ 1" },
              { value: 2, label: "ภาคเรียนที่ 2" },
            ]}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">วันที่เริ่ม</label>
          <DatePicker
            format="YYYY-MM-DD"
            value={startDate ? dayjs(startDate) : null}
            onChange={(d) => setStartDate(d ? d.format("YYYY-MM-DD") : "")}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-slate-500">วันที่จบ</label>
          <DatePicker
            format="YYYY-MM-DD"
            value={endDate ? dayjs(endDate) : null}
            onChange={(d) => setEndDate(d ? d.format("YYYY-MM-DD") : "")}
          />
        </div>
        <Button type="primary" loading={createMutation.isPending} onClick={handleAdd}>
          เพิ่มภาคเรียน
        </Button>
      </div>
    </SectionCard>
  );
}

export function AcademicManager(): ReactNode {
  return (
    <div className="flex flex-col gap-5">
      <YearsSection />
      <SemestersSection />
    </div>
  );
}
