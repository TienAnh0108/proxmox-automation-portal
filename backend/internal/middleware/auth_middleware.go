package middleware

import (
	"net/http"
	"strings"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRole     = "role"
)

// AuthMiddleware trả về 1 Gin middleware kiểm tra JWT access token
// trong header Authorization. Nhận authService qua tham số (dependency
// injection) thay vì tạo mới bên trong — giúp middleware này test được
// độc lập bằng cách truyền vào 1 fake/mock AuthService.
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "thiếu Authorization header"})
			return
		}

		// Header đúng chuẩn phải có dạng "Bearer <token>"
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header sai định dạng, cần 'Bearer <token>'"})
			return
		}

		claims, err := authService.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token không hợp lệ hoặc đã hết hạn"})
			return
		}

		// Đưa thông tin user vào context — Handler phía sau lấy ra dùng,
		// không cần parse token lại lần nữa.
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)

		c.Next()
	}
}
