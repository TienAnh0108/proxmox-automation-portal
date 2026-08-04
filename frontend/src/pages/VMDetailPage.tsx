import { useEffect, useState, useCallback } from "react";
import { useParams } from "react-router-dom";
import { apiFetch } from "@/api/client";
import type { VMDetail } from "@/types/vm";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { VMActions } from "@/components/vm/VMActions";
import { useVMMetrics } from "@/hooks/useVMMetrics";

export function VMDetailPage() {
  const { node, vmid } = useParams<{ node: string; vmid: string }>();
  const [vm, setVm] = useState<VMDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDetail = useCallback(async () => {
    if (!node || !vmid) return;
    try {
      const data = await apiFetch<VMDetail>(`/nodes/${node}/vms/${vmid}`);
      setVm(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Không thể tải chi tiết VM";
      setError(message);
    } finally {
      setIsLoading(false);
    }
  }, [node, vmid]);

  useEffect(() => {
    setIsLoading(true);
    setError(null);
    fetchDetail();
  }, [fetchDetail]);

  useVMMetrics(node ?? "", vmid ?? "", (updatedVm) => setVm(updatedVm));

  if (isLoading) return <p className="text-muted-foreground">Đang tải chi tiết VM...</p>;
  if (error) return <p className="text-red-500">Lỗi: {error}</p>;
  if (!vm || !node || !vmid) return <p className="text-muted-foreground">Không tìm thấy VM.</p>;

  return (
    <div className="max-w-2xl space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <span>
              {vm.name} <span className="text-muted-foreground">#{vm.vmid}</span>
            </span>
            <span className={vm.status === "running" ? "text-green-500" : "text-muted-foreground"}>
              {vm.status}
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4 text-sm text-muted-foreground">
          <div>
            <p className="text-muted-foreground/70">Uptime</p>
            <p className="text-foreground">{formatUptime(vm.uptime_seconds)}</p>
          </div>
          <div>
            <p className="text-muted-foreground/70">Cores</p>
            <p className="text-foreground">{vm.cores}</p>
          </div>
          <div>
            <p className="text-muted-foreground/70">CPU</p>
            <p className="text-foreground">{vm.cpu_percent}%</p>
          </div>
          <div>
            <p className="text-muted-foreground/70">RAM</p>
            <p className="text-foreground">
              {vm.mem_percent}% ({vm.mem_gib} / {vm.maxmem_gib} GiB)
            </p>
          </div>
          <div>
            <p className="text-muted-foreground/70">Disk</p>
            <p className="text-foreground">{vm.maxdisk_gib} GiB</p>
          </div>
        </CardContent>
      </Card>

      <VMActions node={node} vmid={Number(vmid)} onActionComplete={fetchDetail} />
    </div>
  );
}

function formatUptime(seconds: number): string {
  if (seconds === 0) return "—";
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours}h ${minutes}m`;
}