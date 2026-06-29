"use client";

import {
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  PaperClipOutlined,
  ProfileOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import {
  App,
  Button,
  DatePicker,
  Input,
  Modal,
  Popconfirm,
  Select,
  Spin,
  Tag,
  Upload,
} from "antd";
import dayjs from "dayjs";
import { useState, type ReactNode } from "react";

import { SectionCard } from "@/shared/ui/SectionCard";
import {
  workFileTypeLabel,
  type PersonnelWorkRecord,
  type WorkBody,
  type WorkFileType,
} from "@/shared/schemas/personnelWork";

import {
  useCreateWork,
  useDeleteWork,
  useDeleteWorkFile,
  useUpdateWork,
  useUploadWorkFile,
  useWorkFiles,
  useWorks,
} from "./workHooks";

function dateText(s: string): string {
  return s === "" ? "—" : s;
}

function formatBytes(n: number): string {
  if (n <= 0) return "—";
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

function isWorkFileType(v: string): v is WorkFileType {
  return v === "image" || v === "document" || v === "certificate";
}

function EmptyHint({ children }: { children: ReactNode }): ReactNode {
  return (
    <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 py-8 text-center text-sm text-slate-400">
      {children}
    </div>
  );
}

function AddBar({ children }: { children: ReactNode }): ReactNode {
  return (
    <div className="mt-4 rounded-xl border border-dashed border-slate-200 bg-slate-50/70 p-4">
      {children}
    </div>
  );
}

// ===== ฟอร์มกรอกข้อมูลผลงาน (ใช้ทั้งเพิ่มและแก้ไข) =====
function WorkFields(props: {
  value: WorkBody;
  onChange: (v: WorkBody) => void;
  titleError?: string;
}): ReactNode {
  const { value, onChange, titleError } = props;
  return (
    <div className="flex flex-col gap-3">
      <div>
        <label className="mb-1 block text-xs text-slate-500">ชื่อผลงาน</label>
        <Input
          value={value.title}
          onChange={(e) => onChange({ ...value, title: e.target.value })}
          placeholder="เช่น รางวัลครูดีเด่น, ผลงานวิจัยในชั้นเรียน"
          status={titleError ? "error" : ""}
        />
        {titleError && <p className="mt-1 text-sm text-red-500">{titleError}</p>}
      </div>
      <div>
        <label className="mb-1 block text-xs text-slate-500">รายละเอียด</label>
        <Input.TextArea
          value={value.description}
          onChange={(e) => onChange({ ...value, description: e.target.value })}
          rows={2}
          placeholder="รายละเอียดเพิ่มเติม (ไม่บังคับ)"
        />
      </div>
      <div>
        <label className="mb-1 block text-xs text-slate-500">วันที่ของผลงาน</label>
        <DatePicker
          format="YYYY-MM-DD"
          value={value.work_date ? dayjs(value.work_date) : null}
          onChange={(d) => onChange({ ...value, work_date: d ? d.format("YYYY-MM-DD") : "" })}
        />
      </div>
    </div>
  );
}

const emptyWork: WorkBody = { title: "", description: "", work_date: "" };

// ===== ฟอร์มเพิ่มผลงาน =====
export function WorkAddForm(props: { onAdd: (body: WorkBody) => void; submitting: boolean }): ReactNode {
  const { onAdd, submitting } = props;
  const [value, setValue] = useState<WorkBody>(emptyWork);
  const [error, setError] = useState("");

  const submit = (): void => {
    if (value.title.trim() === "") {
      setError("กรุณากรอกชื่อผลงาน");
      return;
    }
    setError("");
    onAdd({ ...value, title: value.title.trim() });
    setValue(emptyWork);
  };

  return (
    <div className="flex flex-col gap-3">
      <WorkFields value={value} onChange={setValue} titleError={error} />
      <div>
        <Button type="primary" loading={submitting} onClick={submit}>
          เพิ่มผลงาน
        </Button>
      </div>
    </div>
  );
}

// ===== Modal แก้ไขผลงาน =====
function WorkEditModal(props: {
  open: boolean;
  initial: WorkBody;
  submitting: boolean;
  onSave: (body: WorkBody) => void;
  onClose: () => void;
}): ReactNode {
  const { open, initial, submitting, onSave, onClose } = props;
  const [value, setValue] = useState<WorkBody>(initial);
  const [error, setError] = useState("");

  // sync ค่าเริ่มต้นเมื่อเปิด modal ของรายการใหม่
  const [lastInitial, setLastInitial] = useState(initial);
  if (open && lastInitial !== initial) {
    setLastInitial(initial);
    setValue(initial);
    setError("");
  }

  const save = (): void => {
    if (value.title.trim() === "") {
      setError("กรุณากรอกชื่อผลงาน");
      return;
    }
    setError("");
    onSave({ ...value, title: value.title.trim() });
  };

  return (
    <Modal
      open={open}
      title="แก้ไขผลงาน"
      onCancel={onClose}
      onOk={save}
      okText="บันทึก"
      cancelText="ยกเลิก"
      confirmLoading={submitting}
      destroyOnHidden
    >
      <WorkFields value={value} onChange={setValue} titleError={error} />
    </Modal>
  );
}

// ===== ตัวจัดการไฟล์แนบของผลงานหนึ่งรายการ =====
function WorkFilesManager(props: { personnelId: string; workId: string }): ReactNode {
  const { personnelId, workId } = props;
  const { message } = App.useApp();
  const { data, isLoading } = useWorkFiles(personnelId, workId, true);
  const uploadMutation = useUploadWorkFile(personnelId, workId);
  const deleteMutation = useDeleteWorkFile(personnelId, workId);
  const [fileType, setFileType] = useState<WorkFileType>("certificate");

  const handleUpload = (file: File): void => {
    uploadMutation.mutate(
      { fileType, file },
      {
        onSuccess: () => message.success("อัปโหลดไฟล์แล้ว"),
        onError: (err) => message.error(err.message),
      },
    );
  };

  return (
    <div className="mt-3 rounded-xl bg-slate-50/70 p-3">
      {isLoading ? (
        <Spin size="small" />
      ) : (data?.length ?? 0) === 0 ? (
        <p className="py-2 text-center text-xs text-slate-400">ยังไม่มีไฟล์แนบ</p>
      ) : (
        <ul className="flex flex-col gap-1">
          {(data ?? []).map((f) => (
            <li
              key={f.id}
              className="flex items-center justify-between gap-2 rounded-lg bg-white px-3 py-2"
            >
              <div className="flex min-w-0 flex-wrap items-center gap-2">
                <Tag color="blue" bordered={false}>
                  {isWorkFileType(f.file_type) ? workFileTypeLabel[f.file_type] : f.file_type}
                </Tag>
                <span className="truncate text-sm text-slate-700">{f.original_name || "ไฟล์"}</span>
                <span className="num text-xs text-slate-400">{formatBytes(f.size_bytes)}</span>
              </div>
              <div className="flex shrink-0 items-center gap-1">
                {/* ดาวน์โหลดผ่าน signed URL ที่ backend ออกให้ (หมดอายุ) */}
                <a href={f.url} target="_blank" rel="noreferrer">
                  <Button type="text" size="small" icon={<DownloadOutlined />} aria-label="ดาวน์โหลด" />
                </a>
                <Popconfirm
                  title="ลบไฟล์นี้?"
                  okText="ลบ"
                  cancelText="ยกเลิก"
                  okButtonProps={{ danger: true }}
                  onConfirm={() =>
                    deleteMutation.mutate(f.id, {
                      onSuccess: () => message.success("ลบไฟล์แล้ว"),
                      onError: (err) => message.error(err.message),
                    })
                  }
                >
                  <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบไฟล์" />
                </Popconfirm>
              </div>
            </li>
          ))}
        </ul>
      )}

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Select<WorkFileType>
          value={fileType}
          onChange={setFileType}
          size="small"
          style={{ width: 130 }}
          options={[
            { value: "certificate", label: workFileTypeLabel.certificate },
            { value: "document", label: workFileTypeLabel.document },
            { value: "image", label: workFileTypeLabel.image },
          ]}
        />
        <Upload
          showUploadList={false}
          beforeUpload={(file) => {
            handleUpload(file);
            return false; // กัน antd อัปโหลดเอง — เรียกผ่าน mutation ของเรา
          }}
        >
          <Button size="small" icon={<UploadOutlined />} loading={uploadMutation.isPending}>
            อัปโหลดไฟล์
          </Button>
        </Upload>
        <span className="text-xs text-slate-400">รองรับไฟล์สูงสุด 10 MB</span>
      </div>
    </div>
  );
}

// ===== รายการผลงานหนึ่งแถว =====
function WorkRow(props: {
  personnelId: string;
  work: PersonnelWorkRecord;
  onEdit: (work: PersonnelWorkRecord) => void;
}): ReactNode {
  const { personnelId, work, onEdit } = props;
  const { message } = App.useApp();
  const deleteMutation = useDeleteWork(personnelId);
  const [showFiles, setShowFiles] = useState(false);

  return (
    <li className="py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-medium text-slate-700">{work.title}</span>
            <span className="text-xs text-slate-400">
              วันที่ <span className="num">{dateText(work.work_date)}</span>
            </span>
          </div>
          {work.description !== "" && (
            <p className="mt-0.5 text-sm text-slate-500">{work.description}</p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            aria-label="แก้ไข"
            onClick={() => onEdit(work)}
          />
          <Popconfirm
            title="ลบผลงานนี้และไฟล์แนบทั้งหมด?"
            okText="ลบ"
            cancelText="ยกเลิก"
            okButtonProps={{ danger: true }}
            onConfirm={() =>
              deleteMutation.mutate(work.id, {
                onSuccess: () => message.success("ลบผลงานแล้ว"),
                onError: (err) => message.error(err.message),
              })
            }
          >
            <Button type="text" size="small" danger icon={<DeleteOutlined />} aria-label="ลบ" />
          </Popconfirm>
        </div>
      </div>

      <Button
        type="link"
        size="small"
        className="!px-0"
        icon={<PaperClipOutlined />}
        onClick={() => setShowFiles((v) => !v)}
      >
        {showFiles ? "ซ่อนไฟล์แนบ" : `ไฟล์แนบ (${work.file_count})`}
      </Button>
      {showFiles && <WorkFilesManager personnelId={personnelId} workId={work.id} />}
    </li>
  );
}

// ===== ผลงานครู (container) =====
export function WorksSection({ personnelId }: { personnelId: string }): ReactNode {
  const { message } = App.useApp();
  const { data, isLoading } = useWorks(personnelId);
  const createMutation = useCreateWork(personnelId);
  const updateMutation = useUpdateWork(personnelId);
  const [editing, setEditing] = useState<PersonnelWorkRecord | null>(null);

  const handleAdd = (body: WorkBody): void => {
    createMutation.mutate(body, {
      onSuccess: () => message.success("เพิ่มผลงานแล้ว"),
      onError: (err) => message.error(err.message),
    });
  };

  const handleSaveEdit = (body: WorkBody): void => {
    if (!editing) return;
    updateMutation.mutate(
      { workId: editing.id, body },
      {
        onSuccess: () => {
          message.success("บันทึกผลงานแล้ว");
          setEditing(null);
        },
        onError: (err) => message.error(err.message),
      },
    );
  };

  return (
    <SectionCard
      icon={<ProfileOutlined />}
      title="ผลงาน (รายเทอม)"
      description="ผลงานของบุคลากรในเทอมปัจจุบัน แนบไฟล์/เกียรติบัตรได้"
      accent="emerald"
    >
      {isLoading ? (
        <Spin />
      ) : (data?.length ?? 0) === 0 ? (
        <EmptyHint>ยังไม่มีผลงานในเทอมนี้</EmptyHint>
      ) : (
        <ul className="divide-y divide-slate-100">
          {(data ?? []).map((w) => (
            <WorkRow key={w.id} personnelId={personnelId} work={w} onEdit={setEditing} />
          ))}
        </ul>
      )}

      <AddBar>
        <WorkAddForm onAdd={handleAdd} submitting={createMutation.isPending} />
      </AddBar>

      <WorkEditModal
        open={editing !== null}
        initial={
          editing
            ? { title: editing.title, description: editing.description, work_date: editing.work_date }
            : emptyWork
        }
        submitting={updateMutation.isPending}
        onSave={handleSaveEdit}
        onClose={() => setEditing(null)}
      />
    </SectionCard>
  );
}
