// context/AuthContext.tsx
import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { apiFetch, registerTokenHandlers } from "@/api/client";
import type { UserResponse, LoginResponse } from "@/types/auth";

interface AuthContextType {
  user: UserResponse | null;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const accessTokenRef = useRef<string | null>(null);

  // "Cắm" 2 hàm đọc/ghi token thật vào client.ts — chỉ chạy 1 lần lúc mount.
  useEffect(() => {
    registerTokenHandlers(
      () => accessTokenRef.current,
      (token) => {
        accessTokenRef.current = token;
      }
    );
  }, []);

  // Khôi phục session lúc app vừa load (F5) — cookie refresh token vẫn còn
  // dù access token trong RAM đã mất.
  useEffect(() => {
    async function restoreSession() {
      try {
        const res = await apiFetch<{ access_token: string }>(
          "/auth/refresh",
          { method: "POST" }
        );
        accessTokenRef.current = res.access_token;

        const me = await apiFetch<UserResponse>("/auth/me");
        setUser(me);
      } catch {
        // Không có cookie hợp lệ — coi như chưa đăng nhập, không phải lỗi
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    }
    restoreSession();
  }, []);

  async function login(username: string, password: string) {
    const res = await apiFetch<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    accessTokenRef.current = res.access_token;
    setUser(res.user);
  }

  async function logout() {
    try {
      await apiFetch("/auth/logout", { method: "POST" });
    } finally {
      // Luôn xóa state phía client dù API logout có lỗi hay không —
      // user bấm logout thì phải thấy mình logout ra ngay.
      accessTokenRef.current = null;
      setUser(null);
    }
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth phải được dùng bên trong AuthProvider");
  }
  return ctx;
}