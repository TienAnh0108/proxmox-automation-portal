package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS cho phép Frontend (chạy ở origin khác — VD localhost:5173 lúc dev)
// gọi API kèm cookie. Bắt buộc chỉ định origin CỤ THỂ (không dùng "*") vì
// trình duyệt từ chối gửi cookie nếu Access-Control-Allow-Origin là wildcard
// khi Allow-Credentials = true.
func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
