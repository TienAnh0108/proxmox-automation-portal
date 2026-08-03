export interface TaskResponse {
  upid: string;
  node: string;
  vmid: number;
  action: string;
  status: "running" | "stopped";
  exit_status?: string;
  success?: boolean;
  created_by: string;
  created_at: string;
}