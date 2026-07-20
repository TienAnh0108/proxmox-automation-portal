package logger

import "go.uber.org/zap"

var Log *zap.Logger

// Init khởi tạo logger 1 lần duy nhất khi start server
func Init(isProduction bool) {
	var err error
	if isProduction {
		Log, err = zap.NewProduction() // log dạng JSON, phù hợp khi đẩy vào hệ thống log tập trung
	} else {
		Log, err = zap.NewDevelopment() // Log dạng text. dễ đọc hơn khi dev
	}
	if err != nil {
		panic("Không thể khởi tạo logger: " + err.Error())
	}
}
