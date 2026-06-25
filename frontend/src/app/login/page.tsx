"use client";

import { ReadOutlined, SafetyOutlined } from "@ant-design/icons";
import { App } from "antd";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";

import { useAuth } from "@/features/auth/AuthContext";
import { LoginForm } from "@/features/auth/LoginForm";
import { useLogin } from "@/features/auth/useLogin";
import type { LoginRequest } from "@/shared/schemas/auth";

export default function LoginPage(): ReactNode {
  const router = useRouter();
  const { message } = App.useApp();
  const { setSession } = useAuth();
  const [errorMessage, setErrorMessage] = useState<string>("");

  const loginMutation = useLogin((result) => {
    setSession(result);
    message.success(`ยินดีต้อนรับ ${result.user.username}`);
    router.push("/");
  });

  const handleSubmit = (values: LoginRequest): void => {
    setErrorMessage("");
    loginMutation.mutate(values, {
      onError: (error) => setErrorMessage(error.message),
    });
  };

  return (
    <main className="flex min-h-dvh bg-slate-50">
      {/* แผงซ้าย — แบรนด์ gradient ฟ้า (signature เดียวกับ dashboard) ซ่อนบนจอเล็ก */}
      <aside className="bg-brand-gradient relative hidden w-[44%] overflow-hidden p-12 text-white lg:flex lg:flex-col lg:justify-between">
        <div className="pointer-events-none absolute -right-16 -top-16 h-64 w-64 rounded-full bg-white/10" />
        <div className="pointer-events-none absolute -bottom-20 -left-10 h-72 w-72 rounded-full bg-brand-bright/20" />

        <div className="relative flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-white/15 ring-1 ring-inset ring-white/25">
            <ReadOutlined className="text-xl" />
          </div>
          <div className="leading-tight">
            <div className="text-lg font-bold tracking-tight">ชุมโค</div>
            <div className="text-xs text-white/70">ระบบบริหารโรงเรียน</div>
          </div>
        </div>

        <div className="relative">
          <h2 className="text-3xl font-bold leading-snug">
            จัดการงานโรงเรียน
            <br />
            ในที่เดียว
          </h2>
          <p className="mt-3 max-w-sm text-sm text-white/75">
            บุคลากร นักเรียน เช็คชื่อ และคะแนนความประพฤติ —
            ครบทุกกลุ่มงานในระบบเดียว ใช้งานง่ายสำหรับครูและผู้บริหาร
          </p>
        </div>

        <div className="relative flex items-center gap-2 text-xs text-white/60">
          <SafetyOutlined />
          <span>ข้อมูลส่วนบุคคลได้รับการคุ้มครองตาม PDPA</span>
        </div>
      </aside>

      {/* แผงขวา — ฟอร์มเข้าสู่ระบบ */}
      <div className="flex flex-1 items-center justify-center p-6">
        <div className="w-full max-w-sm">
          {/* แบรนด์สำหรับจอเล็ก (ไม่มีแผงซ้าย) */}
          <div className="mb-8 flex items-center gap-2.5 lg:hidden">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-brand to-brand-deep text-white">
              <ReadOutlined className="text-lg" />
            </div>
            <div className="leading-tight">
              <div className="text-base font-bold text-slate-800">ชุมโค</div>
              <div className="text-xs text-slate-400">ระบบบริหารโรงเรียน</div>
            </div>
          </div>

          <div className="mb-6">
            <h1 className="text-2xl font-bold tracking-tight text-slate-800">เข้าสู่ระบบ</h1>
            <p className="mt-1 text-sm text-slate-500">
              กรอกข้อมูลบัญชีของคุณเพื่อเริ่มใช้งาน
            </p>
          </div>

          <LoginForm
            onSubmit={handleSubmit}
            isSubmitting={loginMutation.isPending}
            errorMessage={errorMessage}
          />
        </div>
      </div>
    </main>
  );
}
