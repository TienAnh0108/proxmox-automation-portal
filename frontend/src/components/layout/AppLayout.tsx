import type { ReactNode } from "react";
import { Header } from "@/components/layout/Header";

export function AppLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="p-6">{children}</main>
    </div>
  );
}