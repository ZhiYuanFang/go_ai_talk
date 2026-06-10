package ucg

import (
	"context"

	"hello/internal/services/device"
)

// GetAIQuotaDefaultAdmin 经 device internal 读取全局 AI 额度默认。
func GetAIQuotaDefaultAdmin(ctx context.Context) (device.AIQuotaDefaultDTO, error) {
	return device.AIQuotaHTTP().RemoteGetDefaultAdmin(ctx)
}

// UpdateAIQuotaDefaultAdmin 经 device internal 更新全局 AI 额度默认。
func UpdateAIQuotaDefaultAdmin(ctx context.Context, polishLimit, voiceLimit int) (device.AIQuotaDefaultDTO, error) {
	return device.AIQuotaHTTP().RemotePutDefaultAdmin(ctx, polishLimit, voiceLimit)
}

// GetAIQuotaUserOverrideAdmin 读取 wxId override。
func GetAIQuotaUserOverrideAdmin(ctx context.Context, wxID int64) (device.AIQuotaUserOverrideDTO, error) {
	return device.AIQuotaHTTP().RemoteGetUserOverrideAdmin(ctx, wxID)
}

// UpdateAIQuotaUserOverrideAdmin 更新 wxId override。
func UpdateAIQuotaUserOverrideAdmin(ctx context.Context, wxID int64, polishLimit, voiceLimit *int, clearPolish, clearVoice bool) (device.AIQuotaUserOverrideDTO, error) {
	return device.AIQuotaHTTP().RemotePutUserOverrideAdmin(ctx, wxID, polishLimit, voiceLimit, clearPolish, clearVoice)
}
