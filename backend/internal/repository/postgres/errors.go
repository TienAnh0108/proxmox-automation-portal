package postgres

import "strings"

// isUniqueViolation kiểm tra lỗi có phải do vi phạm UNIQUE constraint không.
// Cách đơn giản: kiểm tra chuỗi lỗi — đủ dùng cho MVP. Cách chuẩn hơn
// (dùng pgconn.PgError, so sánh Code == "23505") sẽ chính xác hơn nếu
// sau này cần phân biệt nhiều loại lỗi Postgres khác nhau.
func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
