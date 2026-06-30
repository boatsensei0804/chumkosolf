import type { LoginResult, UserInfo } from "@/shared/schemas/auth";

// เก็บ session ฝั่ง client — เฟสนี้ใช้ localStorage (ภายหลังย้ายเป็น httpOnly cookie ได้)
// แยก logic การเก็บ token ออกจาก component ตามกฎ business-logic-นอก-component
const ACCESS_TOKEN_KEY = "chumkosoft.access_token";
const REFRESH_TOKEN_KEY = "chumkosoft.refresh_token";
const USER_KEY = "chumkosoft.user";

function isBrowser(): boolean {
  return typeof window !== "undefined";
}

export function saveSession(result: LoginResult): void {
  if (!isBrowser()) return;
  window.localStorage.setItem(ACCESS_TOKEN_KEY, result.access_token);
  window.localStorage.setItem(REFRESH_TOKEN_KEY, result.refresh_token);
  window.localStorage.setItem(USER_KEY, JSON.stringify(result.user));
}

export function clearSession(): void {
  if (!isBrowser()) return;
  window.localStorage.removeItem(ACCESS_TOKEN_KEY);
  window.localStorage.removeItem(REFRESH_TOKEN_KEY);
  window.localStorage.removeItem(USER_KEY);
}

export function getAccessToken(): string | null {
  if (!isBrowser()) return null;
  return window.localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getStoredUser(): UserInfo | null {
  if (!isBrowser()) return null;
  const raw = window.localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as UserInfo;
  } catch {
    return null;
  }
}
