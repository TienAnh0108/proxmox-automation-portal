package proxmox

import "math"

const bytesPerGiB = 1024 * 1024 * 1024

// roundToPercent
func roundToPercent(value float64) float64 {
	return math.Round(value*100*100) / 100 // Nhân 100 để thành % rồi Round làm tròn 2 số
}

// calcPercent tính phần trăm từ 2 giá trị used/total (dùng cho mem, disk)
func calcPercent(used, total int64) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(used)/float64(total)*100*100) / 100
}

// Byte to GiB
func bytesToGiB(bytes int64) float64 {
	return math.Round(float64(bytes)/bytesPerGiB*100) / 100
}
