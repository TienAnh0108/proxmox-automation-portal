package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/dto"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/logger"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/repository"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/repository/postgres"
	"go.uber.org/zap"
)

type TaskService interface {
	RecordTask(ctx context.Context, node string, vmid int, action, upid, createdBy string)
	GetTaskStatus(ctx context.Context, upid, requesterID, requesterRole string) (*dto.TaskResponse, error)
}

type taskService struct {
	taskRepo      repository.TaskRepository
	proxmoxClient *proxmox.Client
}

func NewTaskService(taskRepo repository.TaskRepository, proxmoxClient *proxmox.Client) TaskService {
	return &taskService{taskRepo: taskRepo, proxmoxClient: proxmoxClient}
}

// RecordTask lưu task vào DB sau khi action đã submit thành công lên Proxmox.
// Cố ý KHÔNG trả error: VM đã được submit thật rồi, lỗi ghi log task chỉ nên
// log lại chứ không làm fail request action của client (trade-off đã chốt).
func (s *taskService) RecordTask(ctx context.Context, node string, vmid int, action, upid, createdBy string) {
	task := &domain.Task{
		UPID:      upid,
		Node:      node,
		VMID:      vmid,
		Action:    action,
		Status:    domain.TaskStatusRunning,
		CreatedBy: createdBy,
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		logger.Log.Error("không thể lưu task vào DB",
			zap.Error(err), zap.String("upid", upid), zap.String("action", action))
	}
}

func (s *taskService) GetTaskStatus(ctx context.Context, upid, requesterID, requesterRole string) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByUPID(ctx, upid)
	if err != nil {
		if errors.Is(err, postgres.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}

	if task.CreatedBy != requesterID && domain.Role(requesterRole) != domain.RoleAdmin {
		return nil, ErrTaskForbidden
	}

	result, err := s.proxmoxClient.GetTaskStatus(task.Node, task.UPID)
	if err != nil {
		// Không poll được Proxmox (VD: task đã bị Proxmox dọn log cũ) —
		// vẫn trả dữ liệu cũ trong DB thay vì fail cả request.
		logger.Log.Warn("không thể poll trạng thái task từ Proxmox",
			zap.Error(err), zap.String("upid", upid))
		return toTaskResponse(task), nil
	}

	status := domain.TaskStatus(result.Status)
	var exitStatus *string
	if result.ExitStatus != "" {
		exitStatus = &result.ExitStatus
	}

	if err := s.taskRepo.UpdateStatus(ctx, task.UPID, status, exitStatus); err != nil {
		logger.Log.Error("không thể cập nhật trạng thái task", zap.Error(err), zap.String("upid", upid))
	}

	task.Status = status
	task.ExitStatus = exitStatus
	return toTaskResponse(task), nil
}

func toTaskResponse(t *domain.Task) *dto.TaskResponse {
	return &dto.TaskResponse{
		UPID:       t.UPID,
		Node:       t.Node,
		VMID:       t.VMID,
		Action:     t.Action,
		Status:     string(t.Status),
		ExitStatus: t.ExitStatus,
		CreatedBy:  t.CreatedBy,
		CreatedAt:  t.CreatedAt,
	}
}
