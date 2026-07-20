package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Proxmox
	ProxmoxHost        string
	ProxmoxTokenID     string
	ProxmoxTokenSecret string

	// Server
	ServerPort string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() *Config {
	viper.SetConfigFile("internal/config/config.env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Không thể đọc file config.env:", err)
	}

	cfg := &Config{
		ProxmoxHost:        viper.GetString("PROXMOX_HOST"),
		ProxmoxTokenID:     viper.GetString("PROXMOX_API_TOKEN_ID"),
		ProxmoxTokenSecret: viper.GetString("PROXMOX_API_TOKEN_SECRET"),

		ServerPort: viper.GetString("SERVER_PORT"),

		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),

		JWTSecret:       viper.GetString("JWT_SECRET"),
		AccessTokenTTL:  time.Duration(viper.GetInt("ACCESS_TOKEN_TTL_MINUTES")) * time.Minute,
		RefreshTokenTTL: time.Duration(viper.GetInt("REFRESH_TOKEN_TTL_DAYS")) * 24 * time.Hour,
	}

	cfg.validate()
	return cfg
}

// Validate kiểm tra các biến bắt buộc, fail-fast ngay khi khởi động
// thay vì để lỗi xảy ra giữa chừng lúc runtime
func (c *Config) validate() {
	required := map[string]string{
		"PROXMOX_HOST":             c.ProxmoxHost,
		"PROXMOX_API_TOKEN_ID":     c.ProxmoxTokenID,
		"PROXMOX_API_TOKEN_SECRET": c.ProxmoxTokenSecret,
		"DB_HOST":                  c.DBHost,
		"DB_USER":                  c.DBUser,
		"DB_NAME":                  c.DBName,
		"JWT_SECRET":               c.JWTSecret,
	}
	for key, val := range required {
		if val == "" {
			log.Fatalf("Thiếu biến môi trường bắt buộc: %s", key)
		}
	}
	if len(c.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET phải có ít nhất 32 ký tự")
	}
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}
