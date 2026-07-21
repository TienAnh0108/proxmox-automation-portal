package handler

import (
	"errors"
	"net/http"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/dto"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/logger"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/repository/postgres"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	resp, err := h.authService.Refresh(c.Request.Context(), req)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req); err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "đăng xuất thành công"})
}

// Me trả thông tin user hiện tại — dữ liệu lấy từ context, đã được
// AuthMiddleware set sẵn, KHÔNG cần query lại DB.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
		"role":     role,
	})
}

// writeAuthError ánh xạ lỗi nghiệp vụ (từ Service) sang đúng HTTP status
// code. Gom về 1 hàm dùng chung để không lặp lại switch-case ở mọi handler.
func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sai tên đăng nhập hoặc mật khẩu"})
	case errors.Is(err, service.ErrInvalidRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": "role không hợp lệ"})
	case errors.Is(err, service.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token không hợp lệ"})
	case errors.Is(err, service.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token đã hết hạn"})
	case errors.Is(err, service.ErrTokenRevoked):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token đã bị thu hồi"})
	case errors.Is(err, postgres.ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "username đã tồn tại"})
	default:
		// Lỗi không xác định — log ĐẦY ĐỦ chi tiết (bao gồm err gốc) để debug,
		// nhưng response cho client vẫn chung chung, không lộ thông tin nội bộ.
		logger.Log.Error("unhandled auth error",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "đã có lỗi xảy ra, vui lòng thử lại"})
	}
}
