import { ArrowLeftOutlined } from "@ant-design/icons";
import Link from "next/link";
import type { ReactNode } from "react";

// PageHeader — ส่วนหัวของหน้า (ตาม design-system/MASTER §6)
// มีลิงก์ย้อนกลับ (ถ้ามี) + ไอคอนกรอบฟ้า + title/subtitle + actions
export function PageHeader({
  icon,
  title,
  subtitle,
  actions,
  backHref,
  backLabel,
}: {
  icon?: ReactNode;
  title: string;
  subtitle?: string;
  actions?: ReactNode;
  backHref?: string;
  backLabel?: string;
}): ReactNode {
  return (
    <div className="flex flex-col gap-3">
      {backHref && (
        <Link
          href={backHref}
          className="inline-flex w-fit items-center gap-1.5 text-sm text-slate-500 transition-colors hover:text-brand"
        >
          <ArrowLeftOutlined className="text-xs" />
          {backLabel ?? "ย้อนกลับ"}
        </Link>
      )}
      <div className="flex items-end justify-between gap-4">
        <div className="flex items-center gap-3">
          {icon && (
            <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-brand/10 text-xl text-brand">
              {icon}
            </span>
          )}
          <div className="min-w-0">
            <h1 className="text-xl font-bold tracking-tight text-slate-800">{title}</h1>
            {subtitle && <p className="text-sm text-slate-500">{subtitle}</p>}
          </div>
        </div>
        {actions}
      </div>
    </div>
  );
}
