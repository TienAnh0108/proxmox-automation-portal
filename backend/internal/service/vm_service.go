package service

import (
	"context"
	"fmt"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
)

// VMService chứa business rule cho vòng đời VM — cụ thể là kiểm tra trạng
// thái hiện tại trước khi cho phép 1 action chạy, tránh gửi lệnh vô nghĩa
// lên Proxmox (VD: Start 1 VM đang chạy sẵn).
type VMService interface {
	Start(ctx context.Context, node string, vmid int, userID string) (string, error)
	Stop(ctx context.Context, node string, vmid int, userID string) (string, error)
	Shutdown(ctx context.Context, node string, vmid int, userID string) (string, error)
	Reboot(ctx context.Context, node string, vmid int, userID string) (string, error)
	Reset(ctx context.Context, node string, vmid int, userID string) (string, error)
	Delete(ctx context.Context, node string, vmid int, userID string) (string, error)
}

type vmService struct {
	proxmoxClient *proxmox.Client
	taskService   TaskService
}

func NewVMService(proxmoxClient *proxmox.Client, taskService TaskService) VMService {
	return &vmService{proxmoxClient: proxmoxClient, taskService: taskService}
}

// vmActionFunc là chữ ký chung của các hàm action trên proxmox.Client
// (StartVM, StopVM, ShutdownVM, RebootVM, ResetVM, DeleteVM đều cùng dạng).
type vmActionFunc func(node string, vmid int) (string, error)

// executeAction là hàm dùng chung cho mọi action: kiểm tra trạng thái hiện
// tại của VM trước, nếu không khớp requiredStatus thì từ chối ngay (409),
// nếu hợp lệ thì gọi action thật lên Proxmox rồi ghi lại task.
func (s *vmService) executeAction(
	ctx context.Context,
	node string, vmid int, userID string,
	actionName, actionLabel, requiredStatus, conflictLabel string,
	action vmActionFunc,
) (string, error) {
	detail, err := s.proxmoxClient.GetVMDetail(node, vmid)
	if err != nil {
		return "", fmt.Errorf("get vm detail: %w", err)
	}

	if detail.Status != requiredStatus {
		return "", &VMStateError{
			Message: fmt.Sprintf("VM đang %s, không thể %s", conflictLabel, actionLabel),
		}
	}

	upid, err := action(node, vmid)
	if err != nil {
		return "", fmt.Errorf("proxmox action: %w", err)
	}

	s.taskService.RecordTask(ctx, node, vmid, actionName, upid, userID)
	return upid, nil
}

func (s *vmService) Start(ctx context.Context, node string, vmid int, userID string) (string, error) {
	return s.executeAction(ctx, node, vmid, userID,
		"start", "Start", "stopped", "chạy",
		s.proxmoxClient.StartVM)
}

func (s *vmService) Stop(ctx context.Context, node string, vmid int, userID string) (string, error) {
	return s.executeAction(ctx, node, vmid, userID,
		"stop", "Stop", "running", "dừng",
		s.proxmoxClient.StopVM)
}

func (s *vmService) Shutdown(ctx context.Context, node string, vmid int, userID string) (string, error) {
	return s.executeAction(ctx, node, vmid, userID,
		"shutdown", "Shutdown", "running", "dừng",
		s.proxmoxClient.ShutdownVM)
}

func (s *vmService) Reboot(ctx context.Context, node string, vmid int, userID string) (string, error) {
	return s.executeAction(ctx, node, vmid, userID,
		"reboot", "Reboot", "running", "dừng",
		s.proxmoxClient.RebootVM)
}

func (s *vmService) Reset(ctx context.Context, node string, vmid int, userID string) (string, error) {
	return s.executeAction(ctx, node, vmid, userID,
		"reset", "Reset", "running", "dừng",
		s.proxmoxClient.ResetVM)
}

func (s *vmService) Delete(ctx context.Context, node string, vmid int, userID string) (string, error) {
	return s.executeAction(ctx, node, vmid, userID,
		"delete", "Delete", "stopped", "chạy",
		s.proxmoxClient.DeleteVM)
}
