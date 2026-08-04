import { useEffect, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import { apiFetch } from "@/api/client";
import type { VM } from "@/types/vm";
import { CloneVMDialog } from "@/components/vm/CloneVMDialog";

export function VMListInline({ node }: { node: string }) {
  const [vms, setVms] = useState<VM[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchVMs = useCallback(async () => {
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
  }, [node]);

  useEffect(() => {
    fetchVMs();
  }, [fetchVMs]);

  if (isLoading) {
    return <p className="p-4 text-sm text-muted-foreground">Đang tải danh sách VM...</p>;
  }

  return (
    <div className="border-t border-border">
      <div className="flex justify-end p-2">
        <CloneVMDialog node={node} vms={vms} onCloneComplete={fetchVMs} />
      </div>

      {error && <p className="p-4 text-sm text-red-500">Lỗi: {error}</p>}

      {!error && vms.length === 0 && (
        <p className="p-4 text-sm text-muted-foreground">Không có VM nào trên node này.</p>
      )}

      {!error && vms.length > 0 && (
        <div className="divide-y divide-border">
          {vms.map((vm) => (
            <Link
              key={vm.vmid}
              to={`/nodes/${node}/vms/${vm.vmid}`}
              className="flex items-center justify-between px-4 py-3 text-sm hover:bg-accent"
            >
              <div>
                <span className="font-medium text-foreground">{vm.name}</span>
                <span className="ml-2 text-muted-foreground">#{vm.vmid}</span>
              </div>
              <div className="flex items-center gap-4 text-muted-foreground">
                <span>CPU {vm.cpu_percent}%</span>
                <span>RAM {vm.mem_percent}%</span>
                <span className={vm.status === "running" ? "text-green-500" : "text-muted-foreground"}>
                  {vm.status}
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}