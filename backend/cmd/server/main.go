package main

import (
	"log"
	"net"
	"time"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/config"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/database"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/logger"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/repository/postgres"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/router"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.AppEnv)
	defer logger.Log.Sync()

	db, err := database.Connect(cfg.PostgresDSN())
	if err != nil {
		log.Fatal("Không thể kết nối database:", err)
	}
	defer db.Close()

	proxmoxClient := proxmox.NewClient(cfg.ProxmoxHost, cfg.ProxmoxTokenID, cfg.ProxmoxTokenSecret)

	userRepo := postgres.NewUserRepository(db)
	refreshRepo := postgres.NewRefreshTokenRepository(db)
	tokenMgr := service.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authService := service.NewAuthService(userRepo, refreshRepo, tokenMgr)

	r := router.SetupRouter(router.Dependencies{
		ProxmoxClient: proxmoxClient,
		AuthService:   authService,
	})

	ip := getLocalIP()
	logger.Log.Info("server đang khởi động",
		zap.String("port", cfg.ServerPort),
		zap.String("local_url", "http://localhost:"+cfg.ServerPort),
		zap.String("lan_url", "http://"+ip+":"+cfg.ServerPort),
	)

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		logger.Log.Fatal("lỗi khi khởi động server", zap.Error(err))
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

var _ = time.Now
