import { useState } from "react";
import { toast } from "sonner";
import { apiFetch } from "@/api/client";
import { usePollTask } from "@/hooks/usePollTask";
import type { VM } from "@/types/vm";
import type { TaskResponse } from "@/types/task";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Props {
  node: string;
  vms: VM[]; // dùng lại data đã fetch sẵn ở VMListInline, không gọi API mới
  onCloneComplete: () => void;
}

export function CloneVMDialog({ node, vms, onCloneComplete }: Props) {
  const [open, setOpen] = useState(false);
  const [templateVmid, setTemplateVmid] = useState("");
  const [newVmid, setNewVmid] = useState("");
  const [name, setName] = useState("");
  const [fullClone, setFullClone] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [activeUpid, setActiveUpid] = useState<string | null>(null);

  const templates = vms.filter((vm) => vm.is_template);

  usePollTask(activeUpid, (task: TaskResponse) => {
    setActiveUpid(null);
    if (task.success) {
      toast.success(`Clone VM thành công`);
      setOpen(false);
      resetForm();
      onCloneComplete();
    } else {
      toast.error(`Clone thất bại: ${task.exit_status ?? "không rõ nguyên nhân"}`);
    }
  });

  function resetForm() {
    setTemplateVmid("");
    setNewVmid("");
    setName("");
    setFullClone(false);
  }

  async function handleSubmit() {
    if (!templateVmid || !newVmid || !name) {
      toast.error("Vui lòng điền đầy đủ thông tin");
      return;
    }

    setIsSubmitting(true);
    try {
      const res = await apiFetch<{ message: string; task_id: string }>(
        `/nodes/${node}/vms/clone`,
        {
          method: "POST",
          body: JSON.stringify({
            template_vmid: Number(templateVmid),
            new_vmid: Number(newVmid),
            name,
            full_clone: fullClone,
          }),
        }
      );
      setActiveUpid(res.task_id);
      toast.info("Đang clone VM...");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Clone thất bại";
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button size="sm" />}>
        + Clone VM
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Clone VM từ Template</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>Template</Label>
            <Select value={templateVmid} onValueChange={(value) => setTemplateVmid(value ?? "")}>
              <SelectTrigger>
                <SelectValue placeholder="Chọn template..." />
              </SelectTrigger>
              <SelectContent>
                {templates.length === 0 ? (
                  <SelectItem value="none" disabled>
                    Không có template nào trên node này
                  </SelectItem>
                ) : (
                  templates.map((t) => (
                    <SelectItem key={t.vmid} value={String(t.vmid)}>
                      {t.name} (#{t.vmid})
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="new-vmid">VMID mới</Label>
            <Input
              id="new-vmid"
              type="number"
              value={newVmid}
              onChange={(e) => setNewVmid(e.target.value)}
              placeholder="VD: 200"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="vm-name">Tên VM</Label>
            <Input
              id="vm-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="VD: my-new-vm"
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              id="full-clone"
              type="checkbox"
              checked={fullClone}
              onChange={(e) => setFullClone(e.target.checked)}
              className="h-4 w-4"
            />
            <Label htmlFor="full-clone">Full Clone (bỏ tích = Linked Clone)</Label>
          </div>

          <Button
            onClick={handleSubmit}
            disabled={isSubmitting || activeUpid !== null}
            className="w-full"
          >
            {activeUpid ? "Đang clone..." : isSubmitting ? "Đang gửi..." : "Clone"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}