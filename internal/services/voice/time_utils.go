package voice

import "time"

// nowUnixSec 与库表 BIGINT Unix 秒时间列一致。
func nowUnixSec() int64 {
	return time.Now().Unix()
}
