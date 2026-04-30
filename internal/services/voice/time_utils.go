package voice

import "time"

// nowText 返回统一的数据库时间文本格式。
func nowText() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
