package database

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // driver, đăng ký dưới tên "pgx"
	"github.com/jmoiron/sqlx"
)

func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối database: %w", err)
	}

	// Kiểm tra kết nối thật sự hoạt động (Connect chỉ validate DSN, không ping)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database thất bại: %w", err)
	}

	return db, nil
}
