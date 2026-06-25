"use client";

import { Spin } from "antd";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { useAuth } from "./AuthContext";

// path ที่เข้าได้โดยไม่ต้องล็อกอิน
const PUBLIC_PATHS = ["/login"];

function isPublicPath(pathname: string): boolean {
  return PUBLIC_PATHS.includes(pathname);
}

function FullScreenSpinner(): ReactNode {
  return (
    <div className="flex min-h-screen items-center justify-center bg-white">
      <Spin size="large" tip="กำลังโหลด...">
        <div className="h-px w-px" />
      </Spin>
    </div>
  );
}

// AuthGuard บังคับสิทธิ์ฝั่ง client:
// - ยังไม่ล็อกอิน + เข้าหน้าใน → เด้งไป /login
// - ล็อกอินแล้ว + เข้า /login → เด้งกลับหน้าแรก
export function AuthGuard({ children }: { children: ReactNode }): ReactNode {
  const router = useRouter();
  const pathname = usePathname();
  const { isHydrated, isAuthenticated } = useAuth();

  const publicPath = isPublicPath(pathname);

  useEffect(() => {
    if (!isHydrated) return;
    if (!isAuthenticated && !publicPath) {
      router.replace("/login");
    } else if (isAuthenticated && publicPath) {
      router.replace("/");
    }
  }, [isHydrated, isAuthenticated, publicPath, router]);

  // รอ hydrate session ก่อน เพื่อกัน flash ของเนื้อหาที่ไม่ควรเห็น
  if (!isHydrated) {
    return <FullScreenSpinner />;
  }
  // กำลังจะ redirect (ผิดเงื่อนไข) → ยังไม่ render เนื้อหา
  if (!isAuthenticated && !publicPath) {
    return <FullScreenSpinner />;
  }
  if (isAuthenticated && publicPath) {
    return <FullScreenSpinner />;
  }

  return <>{children}</>;
}
