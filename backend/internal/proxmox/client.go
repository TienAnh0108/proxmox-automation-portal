package proxmox

import (
	"crypto/tls"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	host        string
	tokenID     string
	tokenSecret string
	client      *resty.Client
}

// NewClient tạo mới một Proxmox API client
// host: ví dụ "https://192.168.1.10:8006"
// tokenID: ví dụ "root@pam!mytoken"
// tokenSecret: chuỗi secret được Proxmox cấp khi tạo token
func NewClient(host, tokenID, tokenSecret string) *Client {
	restyClient := resty.New().
		SetBaseURL(host+"/api2/json").
		SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", tokenID, tokenSecret)).
		// Bỏ qua kiểm tra chứng chỉ SSL tự ký (self-signed cert) - chỉ dùng cho môi trường dev/test
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &Client{
		host:        host,
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		client:      restyClient,
	}
}
