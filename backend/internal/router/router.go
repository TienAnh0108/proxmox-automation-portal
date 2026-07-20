package router

import (
	"net/http"
	"strconv"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/handler"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/middleware"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	ProxmoxClient *proxmox.Client
	AuthService   service.AuthService
}

func SetupRouter(deps Dependencies) *gin.Engine {
	r := gin.Default()

	authHandler := handler.NewAuthHandler(deps.AuthService)

	api := r.Group("/api")
	{
		// ===== Auth routes — public, không cần token =====
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// ===== Auth routes — cần đã đăng nhập =====
		authProtected := api.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(deps.AuthService))
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.GET("/me", authHandler.Me)
			// Register chỉ admin mới được tạo tài khoản mới cho người khác
			authProtected.POST("/register", middleware.RequireRole(domain.RoleAdmin), authHandler.Register)
		}

		// ===== VM/Node routes — cần đã đăng nhập =====
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(deps.AuthService))
		{
			protected.GET("/nodes", func(c *gin.Context) {
				nodes, err := deps.ProxmoxClient.ListNodes()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, nodes)
			})

			protected.GET("/nodes/:node/vms", func(c *gin.Context) {
				node := c.Param("node")
				vms, err := deps.ProxmoxClient.ListVMs(node)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, vms)
			})

			protected.POST("/nodes/:node/vms", func(c *gin.Context) {
				node := c.Param("node")
				var req proxmox.CreateVMRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
					return
				}
				upid, err := deps.ProxmoxClient.CreateVM(node, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Đang tạo VM", "task_id": upid})
			})

			// Delete VM — chỉ admin mới được xóa
			protected.DELETE("/nodes/:node/vms/:vmid", middleware.RequireRole(domain.RoleAdmin), func(c *gin.Context) {
				node := c.Param("node")
				vmid, err := strconv.Atoi(c.Param("vmid"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "vmid không hợp lệ"})
					return
				}
				upid, err := deps.ProxmoxClient.DeleteVM(node, vmid)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Đang xóa VM", "task_id": upid})
			})

			protected.POST("/nodes/:node/vms/:vmid/start", func(c *gin.Context) {
				handleVMAction(c, deps.ProxmoxClient.StartVM)
			})
			protected.POST("/nodes/:node/vms/:vmid/stop", func(c *gin.Context) {
				handleVMAction(c, deps.ProxmoxClient.StopVM)
			})
			protected.POST("/nodes/:node/vms/:vmid/shutdown", func(c *gin.Context) {
				handleVMAction(c, deps.ProxmoxClient.ShutdownVM)
			})
			protected.POST("/nodes/:node/vms/:vmid/reboot", func(c *gin.Context) {
				handleVMAction(c, deps.ProxmoxClient.RebootVM)
			})
			protected.POST("/nodes/:node/vms/:vmid/reset", func(c *gin.Context) {
				handleVMAction(c, deps.ProxmoxClient.ResetVM)
			})
		}
	}

	return r
}

func handleVMAction(c *gin.Context, action func(node string, vmid int) (string, error)) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vmid không hợp lệ"})
		return
	}
	upid, err := action(node, vmid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lệnh đã được gửi", "task_id": upid})
}
