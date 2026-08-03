export interface Node {
  node: string;
  status: string;
  cpu_percent: number;
  maxcpu: number;
  mem_gib: number;
  maxmem_gib: number;
  mem_percent: number;
  disk_gib: number;
  maxdisk_gib: number;
  disk_percent: number;
}

export interface VM {
  vmid: number;
  name: string;
  status: string;
  cpu_percent: number;
  mem_gib: number;
  maxmem_gib: number;
  mem_percent: number;
  maxdisk_gib: number;
  is_template: boolean;
}

export interface VMDetail extends VM {
  uptime_seconds: number;
  cores: number;
}

export interface CloneVMRequest {
  template_vmid: number;
  new_vmid: number;
  name: string;
  full_clone: boolean;
}