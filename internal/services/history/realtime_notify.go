package history

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// redisChannelAppHistoryNotify App 网关订阅的 Redis Pub/Sub 频道名。
	redisChannelAppHistoryNotify = "app:history:notify"
	redisKeyPieceVerPrefix       = "history:piece:ver:"
	redisKeyPieceDataPrefix      = "history:piece:data:"
)

// bumpPieceCacheEpoch 在 history 变更后递增设备维度版本号，使旧 piece 缓存自然失效。
func bumpPieceCacheEpoch(ctx context.Context, deviceNo string) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	key := redisKeyPieceVerPrefix + deviceNo
	if _, err := g.Redis().Do(ctx, "INCR", key); err != nil {
		glog.Warningf(ctx, "[history-realtime] piece 版本递增失败 deviceNo=%s err=%v", deviceNo, err)
	}
}

// publishHistoryChange 向网关广播历史增删改（尽力而为，失败仅打日志）。
func publishHistoryChange(ctx context.Context, deviceNo, action string, payload interface{}) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return
	}
	body := map[string]interface{}{
		"device_no": deviceNo,
		"action":    action,
		"payload":   payload,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		glog.Warningf(ctx, "[history-realtime] 序列化通知失败 err=%v", err)
		return
	}
	if _, err := g.Redis().Do(ctx, "PUBLISH", redisChannelAppHistoryNotify, string(raw)); err != nil {
		glog.Warningf(ctx, "[history-realtime] PUBLISH 失败 channel=%s err=%v", redisChannelAppHistoryNotify, err)
	}
}

// notifyEventDisplayName 供 App WS 推送：始终使用事件主数据中的标准名称（如「吃奶」），
// 避免 history 行里落的是模型/用户说法等扩展命中名（与 voice 意图里 extra_event_name 同源）。
func notifyEventDisplayName(ctx context.Context, eventID int64, storedRowName string) string {
	stored := strings.TrimSpace(storedRowName)
	if eventID <= 0 {
		return stored
	}
	events, err := ListEventOptions(ctx)
	if err != nil {
		glog.Warningf(ctx, "[history-realtime] 解析标准事件名失败 eventId=%d err=%v，回退为库内 eventName", eventID, err)
		return stored
	}
	for i := range events {
		if events[i].Id == eventID {
			if n := strings.TrimSpace(events[i].Name); n != "" {
				return n
			}
			break
		}
	}
	return stored
}

func historyToNotifyPayload(ctx context.Context, h entity.History) map[string]interface{} {
	payload := map[string]interface{}{
		"id":          h.Id,
		"deviceNo":    h.DeviceNo,
		"eventId":     h.EventId,
		"eventName":   notifyEventDisplayName(ctx, h.EventId, h.EventName),
		"eventNumber": h.EventNumber,
		"startTime":   h.StartTime,
		"endTime":     h.EndTime,
		"remark":      h.Remark,
	}
	if h.PostId > 0 {
		payload["postId"] = h.PostId
	}
	if h.MediaType > 0 {
		payload["mediaType"] = h.MediaType
	}
	if len(h.ImageKeys) > 0 {
		payload["imageKeys"] = h.ImageKeys
	}
	if strings.TrimSpace(h.VideoKey) != "" {
		payload["videoKey"] = h.VideoKey
	}
	return payload
}

func pieceCacheKey(deviceNo string, eventID int64, startTimeUnixSec, endTimeUnixSec, ver int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%d|%d", deviceNo, eventID, startTimeUnixSec, endTimeUnixSec, ver)))
	return redisKeyPieceDataPrefix + hex.EncodeToString(sum[:16])
}
