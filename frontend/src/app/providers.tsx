"use client";

import "@ant-design/v5-patch-for-react-19";

import { AntdRegistry } from "@ant-design/nextjs-registry";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { App as AntdApp, ConfigProvider } from "antd";
import thTH from "antd/locale/th_TH";
import { useState, type ReactNode } from "react";

import { AuthGuard } from "@/features/auth/AuthGuard";
import { AuthProvider } from "@/features/auth/AuthContext";
import { antdTheme } from "@/lib/theme";

// Providers รวม: antd SSR registry + theme (locale ไทย) + App + React Query + auth (session + route guard)
export function Providers({ children }: { children: ReactNode }): ReactNode {
  const [queryClient] = useState(() => new QueryClient());

  return (
    <AntdRegistry>
      <ConfigProvider theme={antdTheme} locale={thTH}>
        <AntdApp>
          <QueryClientProvider client={queryClient}>
            <AuthProvider>
              <AuthGuard>{children}</AuthGuard>
            </AuthProvider>
          </QueryClientProvider>
        </AntdApp>
      </ConfigProvider>
    </AntdRegistry>
  );
}
