import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { apiFetch } from "@/api/client";
import type { VM } from "@/types/vm";

export function VMListInline({ node }: { node: string }) {
  const [vms, setVms] = useState<VM[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchVMs() {
      setIsLoading(true);
      setError(null);
      try {
        const data = await apiFetch<VM[]>(`/nodes/${node}/vms`);
        setVms(data);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Không thể tải danh sách VM";
        setError(message);
      } finally {
        setIsLoading(false);
      }
    }
    fetchVMs();
  }, [node]); // chạy lại nếu prop `node` đổi — dù thực tế Component này bị unmount/remount mỗi lần đóng/mở accordion nên [] cũng đủ, nhưng khai báo đúng dependency là thói quen đúng

  if (isLoading) {
    return <p className="p-4 text-sm text-neutral-400">Đang tải danh sách VM...</p>;
  }

  if (error) {
    return <p className="p-4 text-sm text-red-500">Lỗi: {error}</p>;
  }

  if (vms.length === 0) {
    return <p className="p-4 text-sm text-neutral-500">Không có VM nào trên node này.</p>;
  }

  return (
    <div className="divide-y divide-neutral-800 border-t border-neutral-800">
      {vms.map((vm) => (
        <Link
          key={vm.vmid}
          to={`/nodes/${node}/vms/${vm.vmid}`}
          className="flex items-center justify-between px-4 py-3 text-sm hover:bg-neutral-800"
        >
          <div>
            <span className="font-medium text-white">{vm.name}</span>
            <span className="ml-2 text-neutral-500">#{vm.vmid}</span>
          </div>
          <div className="flex items-center gap-4 text-neutral-400">
            <span>CPU {vm.cpu_percent}%</span>
            <span>RAM {vm.mem_percent}%</span>
            <span className={vm.status === "running" ? "text-green-500" : "text-neutral-500"}>
              {vm.status}
            </span>
          </div>
        </Link>
      ))}
    </div>
  );
}