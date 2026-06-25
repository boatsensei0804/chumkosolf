import type { ReactNode } from "react";

import { ACCENTS, type Accent } from "./accent";

// SectionCard — การ์ดมาตรฐาน (ตาม design-system/MASTER §6)
// header: ไอคอน chip สี accent + title + description + extra; body มี padding
export function SectionCard({
  icon,
  title,
  description,
  extra,
  children,
  className,
  accent = "blue",
}: {
  icon?: ReactNode;
  title: string;
  description?: string;
  extra?: ReactNode;
  children: ReactNode;
  className?: string;
  accent?: Accent;
}): ReactNode {
  return (
    <section className={`rounded-xl border border-slate-200 bg-white ${className ?? ""}`}>
      <header className="flex items-center gap-2.5 border-b border-slate-100 px-4 py-3">
        {icon && (
          <span
            className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ${ACCENTS[accent].chip}`}
          >
            {icon}
          </span>
        )}
        <div className="min-w-0 flex-1">
          <h2 className="text-sm font-semibold text-slate-800">{title}</h2>
          {description && <p className="truncate text-xs text-slate-400">{description}</p>}
        </div>
        {extra}
      </header>
      <div className="p-4">{children}</div>
    </section>
  );
}
