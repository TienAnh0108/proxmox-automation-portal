package router

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/domain"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/dto"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/handler"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/logger"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/middleware"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/service"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	ProxmoxClient      *proxmox.Client
	AuthService        service.AuthService
	TaskService        service.TaskService
	VMService          service.VMService
	AppEnv             string
	FrontendOrigin     string
	CookieSecure       bool
	RefreshTokenMaxAge int
}

func SetupRouter(deps Dependencies) *gin.Engine {
	if deps.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(middleware.CORS(deps.FrontendOrigin))
	r.Use(middleware.RequestLogger(logger.Log))
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)

	authHandler := handler.NewAuthHandler(deps.AuthService, deps.CookieSecure, deps.RefreshTokenMaxAge)
	taskHandler := handler.NewTaskHandler(deps.TaskService)

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

			protected.GET("/tasks/:upid", taskHandler.GetTaskStatus)

			protected.GET("/nodes/:node/vms", func(c *gin.Context) {
				node := c.Param("node")
				vms, err := deps.ProxmoxClient.ListVMs(node)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, vms)
			})

			protected.GET("/nodes/:node/vms/:vmid", func(c *gin.Context) {
				node := c.Param("node")
				vmid, err := strconv.Atoi(c.Param("vmid"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "vmid không hợp lệ"})
					return
				}

				detail, err := deps.ProxmoxClient.GetVMDetail(node, vmid)
				if err != nil {
					if errors.Is(err, proxmox.ErrVMNotFound) {
						c.JSON(http.StatusNotFound, gin.H{"error": "VM không tồn tại"})
						return
					}
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, detail)
			})

			protected.POST("/nodes/:node/vms/clone", func(c *gin.Context) {
				node := c.Param("node")
				var req dto.CloneVMRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
					return
				}

				upid, err := deps.ProxmoxClient.CloneVM(node, req.TemplateVMID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				userID, _ := c.Get(middleware.ContextKeyUserID)
				deps.TaskService.RecordTask(c.Request.Context(), node, req.NewVMID, "clone", upid, userID.(string))

				c.JSON(http.StatusOK, gin.H{"message": "Đang clone VM từ template", "task_id": upid})
			})

			// Delete VM — chỉ admin mới được xóa
			protected.DELETE("/nodes/:node/vms/:vmid", middleware.RequireRole(domain.RoleAdmin), func(c *gin.Context) {
				node := c.Param("node")
				vmid, err := strconv.Atoi(c.Param("vmid"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "vmid không hợp lệ"})
					return
				}

				userID, _ := c.Get(middleware.ContextKeyUserID)
				upid, err := deps.VMService.Delete(c.Request.Context(), node, vmid, userID.(string))
				if err != nil {
					writeVMError(c, err)
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Đang xóa VM", "task_id": upid})
			})

			protected.POST("/nodes/:node/vms/:vmid/start", func(c *gin.Context) {
				handleVMServiceAction(c, deps.VMService.Start)
			})
			protected.POST("/nodes/:node/vms/:vmid/stop", func(c *gin.Context) {
				handleVMServiceAction(c, deps.VMService.Stop)
			})
			protected.POST("/nodes/:node/vms/:vmid/shutdown", func(c *gin.Context) {
				handleVMServiceAction(c, deps.VMService.Shutdown)
			})
			protected.POST("/nodes/:node/vms/:vmid/reboot", func(c *gin.Context) {
				handleVMServiceAction(c, deps.VMService.Reboot)
			})
			protected.POST("/nodes/:node/vms/:vmid/reset", func(c *gin.Context) {
				handleVMServiceAction(c, deps.VMService.Reset)
			})
		}
	}

	return r
}

type vmServiceAction func(ctx context.Context, node string, vmid int, userID string) (string, error)

func handleVMServiceAction(c *gin.Context, action vmServiceAction) {
	node := c.Param("node")
	vmid, err := strconv.Atoi(c.Param("vmid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vmid không hợp lệ"})
		return
	}

	userID, _ := c.Get(middleware.ContextKeyUserID)
	upid, err := action(c.Request.Context(), node, vmid, userID.(string))
	if err != nil {
		writeVMError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lệnh đã được gửi", "task_id": upid})
}

// writeVMError ánh xạ lỗi từ VMService sang đúng HTTP status — cùng pattern
// với writeAuthError trong auth_handler.go.
func writeVMError(c *gin.Context, err error) {
	var stateErr *service.VMStateError
	if errors.As(err, &stateErr) {
		c.JSON(http.StatusConflict, gin.H{"error": stateErr.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
