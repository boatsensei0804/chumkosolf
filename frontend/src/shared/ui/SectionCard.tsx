import type { ReactNode } from "react";

// SectionCard — การ์ดมาตรฐาน (ตาม design-system/MASTER §6)
// header: ไอคอนกรอบฟ้า + title + description + extra ฝั่งขวา; body มี padding
export function SectionCard({
  icon,
  title,
  description,
  extra,
  children,
  className,
}: {
  icon?: ReactNode;
  title: string;
  description?: string;
  extra?: ReactNode;
  children: ReactNode;
  className?: string;
}): ReactNode {
  return (
    <section className={`rounded-2xl border border-slate-200 bg-white ${className ?? ""}`}>
      <header className="flex items-center gap-3 border-b border-slate-100 px-5 py-4">
        {icon && (
          <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-brand/10 text-brand">
            {icon}
          </span>
        )}
        <div className="min-w-0 flex-1">
          <h2 className="text-base font-semibold text-slate-800">{title}</h2>
          {description && <p className="truncate text-xs text-slate-400">{description}</p>}
        </div>
        {extra}
      </header>
      <div className="p-5">{children}</div>
    </section>
  );
}
