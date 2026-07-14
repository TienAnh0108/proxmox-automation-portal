package main

import (
	"log"
	"net"

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

	port := viper.GetString("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	ip := getLocalIP()
	log.Printf("Server đang chạy tại http://localhost:%s (trên VM) hoặc http://%s:%s (từ máy khác)\n", port, ip, port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Lỗi khi khởi động server:", err)
	}
}

// Lấy địa chỉ nội bộ của VM để đưa vào log
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
