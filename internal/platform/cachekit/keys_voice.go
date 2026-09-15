package cachekit

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// voiceClinicRatePrefix 胖宝 Clinic WS 限流；会话/摘要键已废弃，不再提供 builder。
	voiceClinicRatePrefix     = "voice:clinic:rate:"
	voiceSessionDefaultPrefix = "voice:session:"
)

// VoiceSessionKeyPrefix 可通过 VOICE_SESSION_REDIS_PREFIX 覆盖。
func VoiceSessionKeyPrefix() string {
	if p := strings.TrimSpace(os.Getenv("VOICE_SESSION_REDIS_PREFIX")); p != "" {
		return p
	}
	return voiceSessionDefaultPrefix
}

// VoiceSessionKey 设备语音会话 JSON；TTL 见 voiceChat session 配置。
func VoiceSessionKey(deviceNo string) string {
	return VoiceSessionKeyPrefix() + strings.TrimSpace(deviceNo)
}

// VoiceGuardRateKey 文本接口分钟桶限流；TTL 90s。
func VoiceGuardRateKey(deviceNo, minuteBucket string) string {
	return fmt.Sprintf("voice:guard:rate:%s:%s", deviceNo, minuteBucket)
}

// VoiceGuardIdemKey 文本幂等窗口；TTL 见 VOICE_IDEMPOTENCY_TTL。
func VoiceGuardIdemKey(deviceNo string, hash uint32) string {
	return fmt.Sprintf("voice:guard:idem:%s:%d", deviceNo, hash)
}

// VoiceClinicRateKey 诊所固定窗口限流计数（answer_done 成功后递增）。
func VoiceClinicRateKey(wxID int64) string {
	return voiceClinicRatePrefix + strconv.FormatInt(wxID, 10)
}

// CareAlertDailyKey 遗留：宝宝日护理留意列表 JSON；identifier = deviceNo:day（已不再用于短路）。
func CareAlertDailyKey(deviceNo, dayYYYYMMDD string) (string, error) {
	id := strings.TrimSpace(deviceNo) + ":" + strings.TrimSpace(dayYYYYMMDD)
	return Key(DomainVoice, "carealert", "daily", id)
}

// CareAlertDailyLockKey 生成 single-flight 分布式锁；键含 wxId+deviceNo。
func CareAlertDailyLockKey(wxID int64, deviceNo string) (string, error) {
	id := fmt.Sprintf("%d:%s", wxID, strings.TrimSpace(deviceNo))
	return Key(DomainVoice, "carealert", "lock", id)
}

// CareAlertDailyUsageKey 值得留意按用户日限计数；identifier = wxId:yyyyMMdd（Asia/Shanghai）。
func CareAlertDailyUsageKey(wxID int64, dayYYYYMMDD string) (string, error) {
	id := fmt.Sprintf("%d:%s", wxID, strings.TrimSpace(dayYYYYMMDD))
	return Key(DomainVoice, "carealert", "usage", id)
}

// GrowthTrajectoryDailyUsageKey 成长轨迹按用户日限计数；identifier = wxId:yyyyMMdd（Asia/Shanghai）。
func GrowthTrajectoryDailyUsageKey(wxID int64, dayYYYYMMDD string) (string, error) {
	id := fmt.Sprintf("%d:%s", wxID, strings.TrimSpace(dayYYYYMMDD))
	return Key(DomainVoice, "growthtraj", "daily", id)
}

// PredictImminentPendingKey 宝宝预测待发生全量列表 JSON（权威在 Redis，无 TTL，靠客户端全量替换；刷库丢失可接受）。
// identifier = deviceNo；跨 voice 同步写与延时消费读。
func PredictImminentPendingKey(deviceNo string) (string, error) {
	return Key(DomainVoice, "predict", "pending", strings.TrimSpace(deviceNo))
}

// PredictImminentSyncLockKey 同宝宝同步短锁，防并发半更新；TTL 由业务 SetNXEX 传入（秒级）。
func PredictImminentSyncLockKey(deviceNo string) (string, error) {
	return Key(DomainVoice, "predict", "synclock", strings.TrimSpace(deviceNo))
}

// PredictImminentPushedKey 同一 (deviceNo,eventId) 五分钟推送去重；TTL=5min；跨 consumer 实例共享。
func PredictImminentPushedKey(deviceNo string, eventID int64) (string, error) {
	id := fmt.Sprintf("%s:%d", strings.TrimSpace(deviceNo), eventID)
	return Key(DomainVoice, "predict", "pushed", id)
}
