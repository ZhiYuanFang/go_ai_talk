// 预测事项临近离线提醒：Redis 待办权威 + 延时 MQ 叫醒 + 推前校验。
// 设计见 openspec/changes/predict-imminent-offline-push。
package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	deviceclient "hello/internal/clients/device"
	pushclient "hello/internal/clients/push"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"
	"hello/internal/shared/mq"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	predictImminentLeadSecondsDefault = 300 // 默认提前 5 分钟；可由 VOICE_PREDICT_IMMINENT_LEAD_SECONDS 覆盖
	predictImminentPushDedupTTL       = 5 * time.Minute
	predictImminentHistoryWindowSec   = 1800
	predictImminentSyncLockTTL        = 5 * time.Second
	predictImminentQueueName          = "voice.predict.imminent.q"
	predictImminentConsumerEnabledEnv = "VOICE_PREDICT_IMMINENT_MQ_CONSUMER_ENABLED"
	predictImminentPrefetchEnv        = "VOICE_PREDICT_IMMINENT_MQ_PREFETCH"
	predictImminentLeadSecondsEnv     = "VOICE_PREDICT_IMMINENT_LEAD_SECONDS"
)

var predictImminentCache = cachekit.Default()

// predictImminentLeadSeconds 读取提前叫醒秒数；非法或未配置时回退默认 300。
// Args: 无（读环境变量 VOICE_PREDICT_IMMINENT_LEAD_SECONDS）
// Returns: 正整数秒；最小按 0 处理（立即发延时消息）。
func predictImminentLeadSeconds() int64 {
	v := strings.TrimSpace(os.Getenv(predictImminentLeadSecondsEnv))
	if v == "" {
		return predictImminentLeadSecondsDefault
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return predictImminentLeadSecondsDefault
	}
	return n
}

// PredictImminentPendingItem Redis / MQ 共用待办项。
type PredictImminentPendingItem struct {
	EventId int64  `json:"eventId"`
	NextAt  int64  `json:"nextAt"`
	Title   string `json:"title,omitempty"`
}

// predictImminentFirePayload 延时消息体；消费时须与 Redis 中 eventId+nextAt 一致。
type predictImminentFirePayload struct {
	DeviceNo string `json:"deviceNo"`
	EventId  int64  `json:"eventId"`
	NextAt   int64  `json:"nextAt"`
	Title    string `json:"title,omitempty"`
}

// SyncPredictImminentPending 全量替换宝宝待办并按条发延时叫醒。
// Args: wxID 触发用户；deviceNo 宝宝；items 全量列表（可空=清空）。
// Returns: 写入条数、错误。
// Side Effects: Redis 写；非空时 PublishDelayed；不主动 cancel 旧 MQ。
func SyncPredictImminentPending(ctx context.Context, wxID int64, deviceNo string, items []PredictImminentPendingItem) (int, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if wxID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	if deviceNo == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	// 绑机校验：会话 wx 当前绑定须与请求 deviceNo 一致。
	bound, err := deviceclient.FetchDeviceNoByWxID(ctx, wxID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(bound) != deviceNo {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "deviceNo 与当前绑机不一致")
	}

	lockKey, err := cachekit.PredictImminentSyncLockKey(deviceNo)
	if err != nil {
		return 0, err
	}
	got, lockErr := predictImminentCache.SetNXEX(ctx, lockKey, "1", predictImminentSyncLockTTL)
	if lockErr != nil {
		glog.Warningf(ctx, "[predict-imminent] synclock err deviceNoLen=%d err=%v", len(deviceNo), lockErr)
	} else if !got {
		return 0, gerror.NewCode(gcode.CodeOperationFailed, "同步繁忙，请稍后重试")
	}
	defer func() { _ = predictImminentCache.Del(ctx, lockKey) }()

	// 规范化：丢弃非法项；同 eventId 保留列表中靠后的（最后写入赢语义在整表替换层）。
	cleaned := make([]PredictImminentPendingItem, 0, len(items))
	seen := make(map[int64]int) // eventId -> index in cleaned
	for _, it := range items {
		if it.EventId <= 0 || it.NextAt <= 0 {
			continue
		}
		it.Title = strings.TrimSpace(it.Title)
		if idx, ok := seen[it.EventId]; ok {
			cleaned[idx] = it
			continue
		}
		seen[it.EventId] = len(cleaned)
		cleaned = append(cleaned, it)
	}

	pendingKey, err := cachekit.PredictImminentPendingKey(deviceNo)
	if err != nil {
		return 0, err
	}
	if len(cleaned) == 0 {
		// 空列表：清空 Redis，不发取消类 MQ（旧消息靠消费校验作废）。
		if delErr := predictImminentCache.Del(ctx, pendingKey); delErr != nil {
			glog.Warningf(ctx, "[predict-imminent] clear pending failed deviceNoLen=%d err=%v", len(deviceNo), delErr)
			return 0, gerror.WrapCode(gcode.CodeInternalError, delErr, "清空待办失败")
		}
		return 0, nil
	}

	raw, _ := json.Marshal(cleaned)
	if setErr := predictImminentCache.Set(ctx, pendingKey, string(raw)); setErr != nil {
		return 0, gerror.WrapCode(gcode.CodeInternalError, setErr, "写入待办失败")
	}

	pub, pubErr := mq.NewHTTPPublisherForDelayed()
	if pubErr != nil {
		return 0, gerror.WrapCode(gcode.CodeInternalError, pubErr, "延时发布器不可用")
	}
	now := time.Now().Unix()
	leadSec := predictImminentLeadSeconds()
	for _, it := range cleaned {
		delayMs := (it.NextAt - leadSec - now) * 1000
		if delayMs < 0 {
			delayMs = 0
		}
		payload := predictImminentFirePayload{
			DeviceNo: deviceNo,
			EventId:  it.EventId,
			NextAt:   it.NextAt,
			Title:    it.Title,
		}
		if err := pub.PublishDelayed(ctx, eventkit.RoutingVoicePredictImminentFire.String(), delayMs, payload); err != nil {
			glog.Warningf(ctx, "[predict-imminent] PublishDelayed failed deviceNoLen=%d eventId=%d err=%v", len(deviceNo), it.EventId, err)
			return len(cleaned), gerror.WrapCode(gcode.CodeInternalError, err, "延时消息发布失败")
		}
	}
	return len(cleaned), nil
}

// loadPredictImminentPending 读取宝宝当前待办。
func loadPredictImminentPending(ctx context.Context, deviceNo string) ([]PredictImminentPendingItem, error) {
	key, err := cachekit.PredictImminentPendingKey(deviceNo)
	if err != nil {
		return nil, err
	}
	val, ok, err := predictImminentCache.Get(ctx, key)
	if err != nil || !ok || strings.TrimSpace(val) == "" {
		return nil, err
	}
	var list []PredictImminentPendingItem
	if err := json.Unmarshal([]byte(val), &list); err != nil {
		return nil, err
	}
	return list, nil
}

// pendingMatches 判断 Redis 是否仍有相同 eventId+nextAt。
func pendingMatches(list []PredictImminentPendingItem, eventID, nextAt int64) bool {
	for _, it := range list {
		if it.EventId == eventID && it.NextAt == nextAt {
			return true
		}
	}
	return false
}

// classifyStalePending 分类 stale 原因，便于生产 Warning 日志排查（不打印完整 deviceNo）。
// Returns: reason（stale_empty_pending / stale_event_absent / stale_next_at_mismatch）、redisNextAt（仅 mismatch 时有意义）。
func classifyStalePending(list []PredictImminentPendingItem, eventID, fireNextAt int64) (reason string, redisNextAt int64) {
	if len(list) == 0 {
		return "stale_empty_pending", 0
	}
	for _, it := range list {
		if it.EventId != eventID {
			continue
		}
		if it.NextAt == fireNextAt {
			// 调用方应先 pendingMatches；此处兜底。
			return "", it.NextAt
		}
		return "stale_next_at_mismatch", it.NextAt
	}
	return "stale_event_absent", 0
}

// handlePredictImminentFire 处理单条延时叫醒；业务路径一律成功返回（Ack），禁止 requeue / 二次 Publish。
func handlePredictImminentFire(ctx context.Context, body []byte) error {
	var msg predictImminentFirePayload
	if err := json.Unmarshal(body, &msg); err != nil {
		glog.Warningf(ctx, "[predict-imminent] skip reason=bad_payload err=%v", err)
		return nil // Ack
	}
	msg.DeviceNo = strings.TrimSpace(msg.DeviceNo)
	if msg.DeviceNo == "" || msg.EventId <= 0 || msg.NextAt <= 0 {
		glog.Warningf(ctx, "[predict-imminent] skip reason=invalid_fire deviceNoLen=%d eventId=%d nextAt=%d",
			len(msg.DeviceNo), msg.EventId, msg.NextAt)
		return nil
	}

	list, err := loadPredictImminentPending(ctx, msg.DeviceNo)
	if err != nil {
		glog.Warningf(ctx, "[predict-imminent] skip reason=load_pending_err deviceNoLen=%d eventId=%d nextAt=%d err=%v",
			len(msg.DeviceNo), msg.EventId, msg.NextAt, err)
		return nil
	}
	if !pendingMatches(list, msg.EventId, msg.NextAt) {
		reason, redisNextAt := classifyStalePending(list, msg.EventId, msg.NextAt)
		glog.Warningf(ctx, "[predict-imminent] skip reason=%s deviceNoLen=%d eventId=%d fireNextAt=%d redisNextAt=%d pendingCount=%d",
			reason, len(msg.DeviceNo), msg.EventId, msg.NextAt, redisNextAt, len(list))
		return nil
	}

	// 五分钟去重：进入发送前占位，防止多实例/重投双推。
	dedupKey, err := cachekit.PredictImminentPushedKey(msg.DeviceNo, msg.EventId)
	if err != nil {
		glog.Warningf(ctx, "[predict-imminent] skip reason=dedup_key_err deviceNoLen=%d eventId=%d err=%v",
			len(msg.DeviceNo), msg.EventId, err)
		return nil
	}
	ok, nxErr := predictImminentCache.SetNXEX(ctx, dedupKey, "1", predictImminentPushDedupTTL)
	if nxErr != nil {
		glog.Warningf(ctx, "[predict-imminent] skip reason=dedup_setnx_err deviceNoLen=%d eventId=%d err=%v",
			len(msg.DeviceNo), msg.EventId, nxErr)
		return nil
	}
	if !ok {
		glog.Warningf(ctx, "[predict-imminent] skip reason=dedup_hit deviceNoLen=%d eventId=%d nextAt=%d",
			len(msg.DeviceNo), msg.EventId, msg.NextAt)
		return nil
	}

	// 30 分钟历史闸：根事件展开叶子后 filter。
	if recent, histErr := predictImminentHistoryExists(ctx, msg.DeviceNo, msg.EventId); histErr != nil {
		// 历史不可达时不推，避免误扰；去重键已占，五分钟内不会再推。
		glog.Warningf(ctx, "[predict-imminent] skip reason=history_check_err deviceNoLen=%d eventId=%d err=%v (dedup occupied)",
			len(msg.DeviceNo), msg.EventId, histErr)
		return nil
	} else if recent {
		glog.Warningf(ctx, "[predict-imminent] skip reason=history_exists deviceNoLen=%d eventId=%d nextAt=%d",
			len(msg.DeviceNo), msg.EventId, msg.NextAt)
		return nil
	}

	wxIDs, _, listErr := deviceclient.RemoteListWxIDsByDeviceNo(ctx, msg.DeviceNo)
	if listErr != nil {
		glog.Warningf(ctx, "[predict-imminent] skip reason=list_wx_err deviceNoLen=%d eventId=%d err=%v",
			len(msg.DeviceNo), msg.EventId, listErr)
		return nil
	}
	if len(wxIDs) == 0 {
		glog.Warningf(ctx, "[predict-imminent] skip reason=no_bound_wx deviceNoLen=%d eventId=%d nextAt=%d",
			len(msg.DeviceNo), msg.EventId, msg.NextAt)
		return nil
	}
	// 事件名取客户端 pending title（父事件展示名）；空则不推，避免尴尬占位文案。
	title := strings.TrimSpace(msg.Title)
	if title == "" {
		glog.Warningf(ctx, "[predict-imminent] skip reason=empty_title deviceNoLen=%d eventId=%d nextAt=%d",
			len(msg.DeviceNo), msg.EventId, msg.NextAt)
		return nil
	}
	// 昵称经 device 画像；空或失败回落「宝宝」。
	nick := "宝宝"
	if profile, profErr := DeviceProfile().GetProfile(ctx, msg.DeviceNo); profErr != nil {
		glog.Warningf(ctx, "[predict-imminent] profile_err deviceNoLen=%d eventId=%d err=%v (fallback nick=宝宝)",
			len(msg.DeviceNo), msg.EventId, profErr)
	} else if n := strings.TrimSpace(profile.BabyName); n != "" {
		nick = n
	}
	// 可见正文：{昵称}要{title}了；厂商通知标题仍为「胖宝」。
	alert := nick + "要" + title + "了"
	data := map[string]string{
		"bizType":  pushclient.BizPredictImminent,
		"deviceNo": msg.DeviceNo,
		"eventId":  strconv.FormatInt(msg.EventId, 10),
		"nextAt":   strconv.FormatInt(msg.NextAt, 10),
	}
	glog.Infof(ctx, "[predict-imminent] push_attempt deviceNoLen=%d eventId=%d nextAt=%d wxCount=%d",
		len(msg.DeviceNo), msg.EventId, msg.NextAt, len(wxIDs))
	for _, wxID := range wxIDs {
		// badge=0：预测临近不累计 UCG 未读角标。
		if pushErr := pushclient.PushByBizType(ctx, wxID, pushclient.BizPredictImminent, alert, 0, false, data); pushErr != nil {
			// 推送失败只打日志，不 return err（避免 Nack requeue 风暴）。
			glog.Warningf(ctx, "[predict-imminent] push_failed wxId=%d eventId=%d err=%v", wxID, msg.EventId, pushErr)
		}
	}
	return nil
}

// predictImminentHistoryExists 近 30 分钟是否已有同类 history（根则含后代叶子）。
func predictImminentHistoryExists(ctx context.Context, deviceNo string, eventID int64) (bool, error) {
	ids, err := expandEventIDsForHistory(ctx, eventID)
	if err != nil {
		return false, err
	}
	now := time.Now().Unix()
	start := now - predictImminentHistoryWindowSec
	list, err := DeviceHistory().ListHistoryFilter(ctx, deviceNo, ids, start, now, 1, "", false)
	if err != nil {
		return false, err
	}
	return len(list) > 0, nil
}

// expandEventIDsForHistory 将可能为一级根的 eventID 展开为根+全部后代叶子。
func expandEventIDsForHistory(ctx context.Context, rootOrLeaf int64) ([]int64, error) {
	events, err := DeviceAdmin().ListEvents(ctx)
	if err != nil {
		return []int64{rootOrLeaf}, err
	}
	children := make(map[int64][]int64)
	for _, e := range events {
		pid := e.ParentId
		if pid < 0 {
			pid = 0
		}
		children[pid] = append(children[pid], e.Id)
	}
	out := []int64{rootOrLeaf}
	queue := []int64{rootOrLeaf}
	seen := map[int64]struct{}{rootOrLeaf: {}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, c := range children[cur] {
			if _, ok := seen[c]; ok {
				continue
			}
			seen[c] = struct{}{}
			out = append(out, c)
			queue = append(queue, c)
		}
	}
	return out, nil
}

// StartPredictImminentMQConsumer voice-service 启动延时叫醒 consumer（AMQP push，非 ticker）。
// 环境变量 VOICE_PREDICT_IMMINENT_MQ_CONSUMER_ENABLED 默认 true；false 时不订阅。
func StartPredictImminentMQConsumer(ctx context.Context) {
	if !predictImminentConsumerEnabled() {
		glog.Infof(ctx, "[predict-imminent] consumer disabled by %s", predictImminentConsumerEnabledEnv)
		return
	}
	amqpURL, err := predictImminentAMQPURL()
	if err != nil {
		glog.Warningf(ctx, "[predict-imminent] AMQP disabled: %v", err)
		return
	}
	eventkit.RunSharedAMQPConsumers(ctx, eventkit.SharedAMQPConfig{
		URL: amqpURL,
		Subscriptions: []eventkit.AMQPQueueSubscription{{
			QueueName:         predictImminentQueueName,
			Prefetch:          predictImminentPrefetch(),
			ConsumerTagPrefix: "voice-predict-imminent",
			Handler: func(ctx context.Context, queueName, routingKey string, body []byte) error {
				_ = queueName
				_ = routingKey
				// 业务路径一律 nil → Ack；禁止 Nack requeue / 消费内二次 PublishDelayed。
				return handlePredictImminentFire(ctx, body)
			},
		}},
	})
	glog.Infof(ctx, "[predict-imminent] AMQP consumer started queue=%s", predictImminentQueueName)
}

func predictImminentConsumerEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(predictImminentConsumerEnabledEnv)))
	if v == "" {
		return true
	}
	return v != "0" && v != "false" && v != "off"
}

func predictImminentPrefetch() int {
	n := 5
	if v := strings.TrimSpace(os.Getenv(predictImminentPrefetchEnv)); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			n = p
		}
	}
	return n
}

func predictImminentAMQPURL() (string, error) {
	if u := strings.TrimSpace(os.Getenv("RABBITMQ_AMQP_URL")); u != "" {
		return u, nil
	}
	host := strings.TrimSpace(os.Getenv("RABBITMQ_HOST"))
	if host == "" {
		host = "rabbitmq"
	}
	port := "5672"
	if v := strings.TrimSpace(os.Getenv("RABBITMQ_AMQP_PORT")); v != "" {
		port = v
	}
	user := strings.TrimSpace(os.Getenv("MQ_USER"))
	if user == "" {
		user = "guest"
	}
	pass := os.Getenv("MQ_PASSWORD")
	if strings.TrimSpace(pass) == "" {
		pass = "guest"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port), nil
}
