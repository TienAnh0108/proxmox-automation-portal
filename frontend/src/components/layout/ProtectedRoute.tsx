// components/layout/ProtectedRoute.tsx
import { Navigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import type { ReactNode } from "react";

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    // Đang chờ khôi phục session — chưa đủ thông tin để quyết định
    // redirect hay không, tạm hiện loading thay vì đá về Login nhầm.
    return <div className="flex h-screen items-center justify-center">Đang tải...</div>;
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}