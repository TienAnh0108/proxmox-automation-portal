import { useEffect, useState } from "react";
import { apiFetch } from "@/api/client";
import type { TaskResponse } from "@/types/task";

const POLL_INTERVAL_MS = 2000;

export function usePollTask(upid: string | null, onDone: (task: TaskResponse) => void) {
  const [task, setTask] = useState<TaskResponse | null>(null);

  useEffect(() => {
    if (!upid) {
      setTask(null);
      return;
    }

    let cancelled = false; // chặn setState nếu Component unmount giữa chừng

    async function poll() {
      try {
        const data = await apiFetch<TaskResponse>(`/tasks/${upid}`);
        if (cancelled) return;

        setTask(data);

        if (data.status === "stopped") {
          onDone(data); // task đã xong, báo cho nơi gọi biết kết quả
        } else {
          setTimeout(poll, POLL_INTERVAL_MS); // chưa xong, hẹn poll lại
        }
      } catch {
        if (!cancelled) {
          setTimeout(poll, POLL_INTERVAL_MS); // lỗi mạng tạm thời, vẫn thử lại
        }
      }
    }

    poll();

    return () => {
      cancelled = true; // cleanup — component unmount hoặc upid đổi
    };
  }, [upid]);

  return task;
}