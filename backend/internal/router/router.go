package api

import (
	"net/http"
	"strconv"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
	"github.com/gin-gonic/gin"
)

func SetupRouter(client *proxmox.Client) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/nodes", func(c *gin.Context) {
			nodes, err := client.ListNodes()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, nodes)
		})

		api.GET("/nodes/:node/vms", func(c *gin.Context) {
			node := c.Param("node")
			vms, err := client.ListVMs(node)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, vms)
		})

		api.POST("/nodes/:node/vms", func(c *gin.Context) {
			node := c.Param("node")

			var req proxmox.CreateVMRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ: " + err.Error()})
				return
			}

			upid, err := client.CreateVM(node, req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Đang tạo VM", "task_id": upid})
		})

		api.DELETE("/nodes/:node/vms/:vmid", func(c *gin.Context) {
			node := c.Param("node")
			vmidStr := c.Param("vmid")

			vmid, err := strconv.Atoi(vmidStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "vmid không hợp lệ"})
				return
			}

			upid, err := client.DeleteVM(node, vmid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Đang xóa VM", "task_id": upid})
		})

		api.POST("/nodes/:node/vms/:vmid/start", func(c *gin.Context) {
			handleVMAction(c, client.StartVM)
		})

		api.POST("/nodes/:node/vms/:vmid/stop", func(c *gin.Context) {
			handleVMAction(c, client.StopVM)
		})

		api.POST("/nodes/:node/vms/:vmid/shutdown", func(c *gin.Context) {
			handleVMAction(c, client.ShutdownVM)
		})

		api.POST("/nodes/:node/vms/:vmid/reboot", func(c *gin.Context) {
			handleVMAction(c, client.RebootVM)
		})

		api.POST("/nodes/:node/vms/:vmid/reset", func(c *gin.Context) {
			handleVMAction(c, client.ResetVM)
		})
	}

	return r
}

// handleVMAction là hàm dùng chung để xử lý các action start/stop/reboot/...
func handleVMAction(c *gin.Context, action func(node string, vmid int) (string, error)) {
	node := c.Param("node")
	vmidStr := c.Param("vmid")

	vmid, err := strconv.Atoi(vmidStr)
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
