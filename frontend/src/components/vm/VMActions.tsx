import { useState } from "react";
import { toast } from "sonner";
import { apiFetch } from "@/api/client";
import { useAuth } from "@/context/AuthContext";
import { usePollTask } from "@/hooks/usePollTask";
import { Button } from "@/components/ui/button";
import type { TaskResponse } from "@/types/task";

type ActionName = "start" | "stop" | "shutdown" | "reboot" | "reset";

interface Props {
  node: string;
  vmid: number;
  onActionComplete: () => void; // gọi lại để fetch lại VM Detail mới nhất
}

const ACTION_LABELS: Record<ActionName, string> = {
  start: "Start",
  stop: "Stop",
  shutdown: "Shutdown",
  reboot: "Reboot",
  reset: "Reset",
};

export function VMActions({ node, vmid, onActionComplete }: Props) {
  const { user } = useAuth();
  const [activeUpid, setActiveUpid] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<ActionName | null>(null);

  usePollTask(activeUpid, (task: TaskResponse) => {
    setActiveUpid(null);
    setPendingAction(null);
    if (task.success) {
      toast.success(`${pendingAction} thành công`);
    } else {
      toast.error(`${pendingAction} thất bại: ${task.exit_status ?? "không rõ nguyên nhân"}`);
    }
    onActionComplete();
  });

  async function runAction(action: ActionName) {
    setPendingAction(action);
    try {
      const res = await apiFetch<{ message: string; task_id: string }>(
        `/nodes/${node}/vms/${vmid}/${action}`,
        { method: "POST" }
      );
      setActiveUpid(res.task_id);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Thao tác thất bại";
      toast.error(message);
      setPendingAction(null);
    }
  }

  async function runDelete() {
    if (!confirm(`Xóa VM #${vmid}? Hành động này không thể hoàn tác.`)) return;
    setPendingAction(null);
    try {
      const res = await apiFetch<{ message: string; task_id: string }>(
        `/nodes/${node}/vms/${vmid}`,
        { method: "DELETE" }
      );
      setActiveUpid(res.task_id);
      toast.info("Đang xóa VM...");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Xóa thất bại";
      toast.error(message);
    }
  }

  const isBusy = activeUpid !== null;

  return (
    <div className="flex flex-wrap gap-2">
      {(Object.keys(ACTION_LABELS) as ActionName[]).map((action) => (
        <Button
          key={action}
          variant="outline"
          size="sm"
          disabled={isBusy}
          onClick={() => runAction(action)}
        >
          {isBusy && pendingAction === action ? "Đang xử lý..." : ACTION_LABELS[action]}
        </Button>
      ))}

      {user?.role === "admin" && (
        <Button variant="destructive" size="sm" disabled={isBusy} onClick={runDelete}>
          Delete
        </Button>
      )}
    </div>
  );
}