package handler

import (
	"errors"
	"net/http"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/logger"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TaskHandler struct {
	taskService service.TaskService
}

func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	upid := c.Param("upid")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	resp, err := h.taskService.GetTaskStatus(
		c.Request.Context(), upid,
		userID.(string), string(role.(domain.Role)),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "task không tồn tại"})
		case errors.Is(err, service.ErrTaskForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "không có quyền xem task này"})
		default:
			logger.Log.Error("unhandled task error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "đã có lỗi xảy ra, vui lòng thử lại"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
