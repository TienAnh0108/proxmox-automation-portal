package main

import (
	"log"

	"github.com/TienAnh0108/proxmox-automation-portal/internal/api"
	"github.com/TienAnh0108/proxmox-automation-portal/internal/proxmox"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("configs/config.env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Không thể đọc file config.env:", err)
	}

	host := viper.GetString("PROXMOX_HOST")
	tokenID := viper.GetString("PROXMOX_API_TOKEN_ID")
	tokenSecret := viper.GetString("PROXMOX_API_TOKEN_SECRET")

	if host == "" || tokenID == "" || tokenSecret == "" {
		log.Fatal("Thiếu biến môi trường PROXMOX_HOST, PROXMOX_API_TOKEN_ID hoặc PROXMOX_API_TOKEN_SECRET")
	}

	client := proxmox.NewClient(host, tokenID, tokenSecret)

	router := api.SetupRouter(client)

	log.Println("Server đang chạy tại http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Lỗi khi khởi động server:", err)
	}
}
