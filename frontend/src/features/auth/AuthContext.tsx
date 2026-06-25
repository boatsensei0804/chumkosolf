"use client";

import { useRouter } from "next/navigation";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import type { LoginResult, UserInfo } from "@/shared/schemas/auth";

import {
  clearSession,
  getAccessToken,
  getStoredUser,
  saveSession,
} from "./storage";

type AuthContextValue = {
  // user ปัจจุบัน (null = ยังไม่ล็อกอิน)
  user: UserInfo | null;
  // true หลังอ่าน session จาก localStorage เสร็จ — กัน flash ก่อนรู้สถานะจริง
  isHydrated: boolean;
  isAuthenticated: boolean;
  // เก็บ session หลัง login สำเร็จ (เรียกจาก login page)
  setSession: (result: LoginResult) => void;
  // ออกจากระบบ: ล้าง session แล้วเด้งไป /login
  signOut: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

// AuthProvider ถือสถานะ session ของทั้งแอป และ hydrate จาก localStorage ตอน mount
export function AuthProvider({ children }: { children: ReactNode }): ReactNode {
  const router = useRouter();
  const [user, setUser] = useState<UserInfo | null>(null);
  const [isHydrated, setIsHydrated] = useState(false);

  // อ่าน session ที่เก็บไว้ครั้งเดียวหลัง mount (localStorage เข้าถึงได้เฉพาะฝั่ง client)
  useEffect(() => {
    const token = getAccessToken();
    setUser(token ? getStoredUser() : null);
    setIsHydrated(true);
  }, []);

  const setSession = useCallback((result: LoginResult) => {
    saveSession(result);
    setUser(result.user);
  }, []);

  const signOut = useCallback(() => {
    clearSession();
    setUser(null);
    router.replace("/login");
  }, [router]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isHydrated,
      isAuthenticated: user !== null,
      setSession,
      signOut,
    }),
    [user, isHydrated, setSession, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// useAuth คืน session context — ต้องอยู่ภายใต้ AuthProvider
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (ctx === null) {
    throw new Error("useAuth ต้องใช้ภายใน <AuthProvider>");
  }
  return ctx;
}
