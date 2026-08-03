import { useEffect, useState } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import { apiFetch } from "@/api/client";
import type { Node } from "@/types/vm";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { VMListInline } from "@/components/vm/VMListInline";

export function DashboardPage() {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedNode, setExpandedNode] = useState<string | null>(null);

  useEffect(() => {
    async function fetchNodes() {
      try {
        const data = await apiFetch<Node[]>("/nodes");
        setNodes(data);
      } catch (err) {
        const message = err instanceof Error ? err.message : "Không thể tải danh sách node";
        setError(message);
      } finally {
        setIsLoading(false);
      }
    }
    fetchNodes();
  }, []);

  if (isLoading) {
    return <p className="text-neutral-400">Đang tải danh sách node...</p>;
  }

  if (error) {
    return <p className="text-red-500">Lỗi: {error}</p>;
  }

  function toggleNode(nodeName: string) {
    setExpandedNode((current) => (current === nodeName ? null : nodeName));
  }

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {nodes.map((node) => {
        const isExpanded = expandedNode === node.node;
        return (
          <Card key={node.node} className="overflow-hidden">
            <button
              onClick={() => toggleNode(node.node)}
              className="w-full text-left"
            >
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <span className="flex items-center gap-2">
                    {node.node}
                    {isExpanded ? (
                      <ChevronUp className="h-4 w-4 text-neutral-500" />
                    ) : (
                      <ChevronDown className="h-4 w-4 text-neutral-500" />
                    )}
                  </span>
                  <span className={node.status === "online" ? "text-green-500" : "text-red-500"}>
                    {node.status}
                  </span>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-1 text-sm text-neutral-400">
                <p>CPU: {node.cpu_percent}% ({node.maxcpu} cores)</p>
                <p>RAM: {node.mem_percent}% ({node.mem_gib} / {node.maxmem_gib} GiB)</p>
                <p>Disk: {node.disk_percent}% ({node.disk_gib} / {node.maxdisk_gib} GiB)</p>
              </CardContent>
            </button>

            {isExpanded && <VMListInline node={node.node} />}
          </Card>
        );
      })}
    </div>
  );
}