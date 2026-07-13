package main

import (
	"fmt"
	"log"

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

	nodes, err := client.ListNodes()
	if err != nil {
		fmt.Println("Lỗi khi lấy danh sách node:", err)
		return
	}
	fmt.Println("Danh sách Node:", nodes)

	if len(nodes) > 0 {
		vms, err := client.ListVMs(nodes[0].Node)
		if err != nil {
			fmt.Println("Lỗi khi lấy danh sách VM:", err)
			return
		}
		fmt.Println("Danh sách VM:", vms)
	}
}
