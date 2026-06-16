package ucg

import (
	"context"

	"hello/internal/services/contracts"
)

// GetAIQuotaDefaultAdmin 读取 ucg 域润笔全局默认（本地 ai_voice_ucg）。
func GetAIQuotaDefaultAdmin(ctx context.Context) (contracts.PolishAIQuotaDefaultDTO, error) {
	return GetPolishAIQuotaDefaultForAdmin(ctx)
}

// UpdateAIQuotaDefaultAdmin 更新润笔全局默认。
func UpdateAIQuotaDefaultAdmin(ctx context.Context, polishLimit int) (contracts.PolishAIQuotaDefaultDTO, error) {
	return UpdatePolishAIQuotaDefaultForAdmin(ctx, polishLimit)
}

// GetAIQuotaUserOverrideAdmin 读取 wxId 润笔 override。
func GetAIQuotaUserOverrideAdmin(ctx context.Context, wxID int64) (contracts.PolishAIQuotaUserOverrideDTO, error) {
	return GetPolishAIQuotaUserOverrideForAdmin(ctx, wxID)
}

// UpdateAIQuotaUserOverrideAdmin 更新 wxId 润笔 override；polishLimit=nil 表示清除。
func UpdateAIQuotaUserOverrideAdmin(ctx context.Context, wxID int64, polishLimit *int) (contracts.PolishAIQuotaUserOverrideDTO, error) {
	return UpdatePolishAIQuotaUserOverrideForAdmin(ctx, wxID, polishLimit)
}
