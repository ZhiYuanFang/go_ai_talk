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
