import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";

export function Header() {
  const { user, logout } = useAuth();

  return (
    <header className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
      <h1 className="text-lg font-semibold text-foreground">Cloud Automation Portal</h1>
      <div className="flex items-center gap-4">
        <span className="text-sm text-muted-foreground">
          {user?.username} <span className="text-muted-foreground/60">({user?.role})</span>
        </span>
        <Button variant="outline" size="sm" onClick={logout}>
          Đăng xuất
        </Button>
      </div>
    </header>
  );
}