package cachekit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const historyPieceVerPrefix = "history:piece:ver:"
const historyPieceDataPrefix = "history:piece:data:"

// HistoryPieceVerKey 设备 piece 缓存版本 epoch；INCR 触发 piece 数据键失效。
func HistoryPieceVerKey(deviceNo string) string {
	return historyPieceVerPrefix + deviceNo
}

// HistoryPieceDataKey 按查询参数与版本 hash 的 piece 列表 JSON 缓存；TTL 60s。
func HistoryPieceDataKey(deviceNo string, eventID int64, startTimeUnixSec, endTimeUnixSec, ver int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d|%d", deviceNo, eventID, startTimeUnixSec, endTimeUnixSec, ver)))
	return historyPieceDataPrefix + hex.EncodeToString(sum[:16])
}
