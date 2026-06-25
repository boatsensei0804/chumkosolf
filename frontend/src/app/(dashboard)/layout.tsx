import type { ReactNode } from "react";

import { DashboardLayout } from "@/features/navigation/DashboardLayout";

// layout ของกลุ่มหน้า authenticated — ครอบด้วย dashboard shell (sidebar + topbar)
// route guard อยู่ชั้น providers แล้ว จึงมั่นใจว่าผู้ใช้ล็อกอินเมื่อมาถึงที่นี่
export default function DashboardGroupLayout({
  children,
}: {
  children: ReactNode;
}): ReactNode {
  return <DashboardLayout>{children}</DashboardLayout>;
}
