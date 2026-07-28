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

const refreshCookieName = "refresh_token"

type AuthHandler struct {
	authService        service.AuthService
	cookieSecure       bool // true khi đã có HTTPS (đọc từ config, xem bước 6)
	refreshTokenMaxAge int  // giây — dùng làm Max-Age của cookie
}

func NewAuthHandler(authService service.AuthService, cookieSecure bool, refreshTokenMaxAge int) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		cookieSecure:       cookieSecure,
		refreshTokenMaxAge: refreshTokenMaxAge,
	}
}

// setRefreshCookie set cookie HttpOnly chứa refresh token — Path giới hạn
// chỉ gửi kèm khi gọi /api/auth/*, không gửi kèm mọi request khác.
func (h *AuthHandler) setRefreshCookie(c *gin.Context, value string, maxAge int) {
	c.SetCookie(refreshCookieName, value, maxAge, "/api/auth", "", h.cookieSecure, true)
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

	resp, rawRefresh, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, rawRefresh, h.refreshTokenMaxAge)
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err != nil || rawRefresh == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "thiếu refresh token"})
		return
	}

	resp, newRawRefresh, err := h.authService.Refresh(c.Request.Context(), rawRefresh)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, newRawRefresh, h.refreshTokenMaxAge)
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshCookieName)
	if err == nil && rawRefresh != "" {
		if err := h.authService.Logout(c.Request.Context(), rawRefresh); err != nil {
			writeAuthError(c, err)
			return
		}
	}

	// Xóa cookie dù có tìm thấy token hay không — client coi như đã logout.
	h.setRefreshCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"message": "đăng xuất thành công"})
}

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
		logger.Log.Error("unhandled auth error", zap.Error(err), zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "đã có lỗi xảy ra, vui lòng thử lại"})
	}
}
