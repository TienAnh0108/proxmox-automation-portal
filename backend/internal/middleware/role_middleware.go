package middleware

import (
	"net/http"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/gin-gonic/gin"
)

// RequireRole trả về middleware chỉ cho phép request đi tiếp nếu role
// của user (đã được AuthMiddleware set vào context) nằm trong danh sách
// allowedRoles. PHẢI đặt SAU AuthMiddleware trong chain — vì cần
// ContextKeyRole đã được set trước đó.
func RequireRole(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get(ContextKeyRole)
		if !exists {
			// Lỗi lập trình (thiếu AuthMiddleware phía trước), không phải
			// lỗi của client — vẫn trả 401 để không lộ chi tiết nội bộ.
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "chưa xác thực"})
			return
		}

		role, ok := roleValue.(domain.Role)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "lỗi hệ thống"})
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "không đủ quyền truy cập"})
	}
}
