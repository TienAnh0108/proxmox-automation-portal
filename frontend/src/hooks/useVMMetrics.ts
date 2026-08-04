// hooks/useVMMetrics.ts
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { apiFetch } from "@/api/client";
import type { VMDetail } from "@/types/vm";

const POLL_INTERVAL_MS = 8000;
const THRESHOLD_PERCENT = 90;

export function useVMMetrics(
  node: string,
  vmid: string,
  onUpdate: (vm: VMDetail) => void
) {
  // Lưu trạng thái "đã cảnh báo chưa" cho từng chỉ số riêng biệt — tránh
  // bắn toast lặp lại liên tục mỗi 8s trong khi vẫn đang vượt ngưỡng.
  const alertedRef = useRef({ cpu: false, mem: false, disk: false });

  useEffect(() => {
    let cancelled = false;
    let timerId: ReturnType<typeof setTimeout>;

    async function poll() {
      try {
        const vm = await apiFetch<VMDetail>(`/nodes/${node}/vms/${vmid}`);
        if (cancelled) return;

        onUpdate(vm);
        checkThreshold("cpu", vm.cpu_percent, "CPU");
        checkThreshold("mem", vm.mem_percent, "RAM");
        // Backend không có disk_percent riêng ở VMDetail, dùng maxdisk_gib
        // — bỏ qua cảnh báo Disk ở đây vì thiếu % thực tế, chỉ cảnh báo CPU/RAM
      } catch {
        // lỗi mạng tạm thời — bỏ qua, thử lại ở lần poll sau
      } finally {
        if (!cancelled) {
          timerId = setTimeout(poll, POLL_INTERVAL_MS);
        }
      }
    }

    function checkThreshold(key: "cpu" | "mem", value: number, label: string) {
      if (value > THRESHOLD_PERCENT) {
        if (!alertedRef.current[key]) {
          toast.warning(`${label} vượt ngưỡng: ${value}%`);
          alertedRef.current[key] = true;
        }
      } else {
        // Về lại mức bình thường — reset để lần sau vượt ngưỡng lại cảnh báo tiếp
        alertedRef.current[key] = false;
      }
    }

    poll();

    return () => {
      cancelled = true;
      clearTimeout(timerId);
    };
  }, [node, vmid]);
}