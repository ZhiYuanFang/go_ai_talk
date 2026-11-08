package push

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const pushLiveActivityTokenTable = "push_live_activity_token"

// Live Activity token 种类（统一单活动后不再按 attributesKind/activityId 分槽）。
const (
	LATokenKindPushToStart = "push_to_start"
	LATokenKindActivity    = "activity"
)

// entityLiveActivityToken LA 专用 token 行（与 push_device 分表）。
// attributes_kind / activity_id 列保留兼容旧库，业务恒写空串，唯一键为 (wx, device_key, token_kind)。
type entityLiveActivityToken struct {
	ID             uint64
	WxID           int64
	DeviceKey      string
	TokenKind      string
	AttributesKind string
	ActivityID     string
	Token          string
	UpdatedAt      int64
}

// RegisterLiveActivityToken 注册/更新 Push-to-Start 或 per-user activity update token。
//
// 业务逻辑：
//  1. MUST NOT 写入 push_device（避免 uk_token 与 device alert token 互顶）；
//  2. 按 (wx, device_key, token_kind) upsert；不再要求 attributesKind/activityId；
//  3. attributes_kind、activity_id 列强制空串（兼容旧表结构）。
//
// Side Effects: 写 push_live_activity_token。
func RegisterLiveActivityToken(ctx context.Context, wxID int64, deviceKey, tokenKind, token string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	deviceKey = strings.TrimSpace(deviceKey)
	tokenKind = strings.TrimSpace(strings.ToLower(tokenKind))
	token = strings.TrimSpace(token)
	if deviceKey == "" || len(deviceKey) > 64 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceKey 无效")
	}
	if token == "" || len(token) > 512 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "token 无效")
	}
	if tokenKind != LATokenKindPushToStart && tokenKind != LATokenKindActivity {
		return gerror.NewCode(gcode.CodeInvalidParameter, "tokenKind 必须是 push_to_start 或 activity")
	}
	now := time.Now().Unix()
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).Data(g.Map{
		"wx_id":           wxID,
		"device_key":      deviceKey,
		"token_kind":      tokenKind,
		"attributes_kind": "",
		"activity_id":     "",
		"token":           token,
		"updated_at":      now,
	}).OnDuplicate("token", "updated_at").Save()
	if err != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "写入 Live Activity token 失败")
	}
	return nil
}

// ListLiveActivityTokensForWx 列出某用户某 tokenKind 的全部 LA token（可多机）。
func ListLiveActivityTokensForWx(ctx context.Context, wxID int64, tokenKind string) ([]entityLiveActivityToken, error) {
	if wxID <= 0 {
		return nil, nil
	}
	tokenKind = strings.TrimSpace(strings.ToLower(tokenKind))
	var rows []entityLiveActivityToken
	err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("token_kind", tokenKind).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// DeleteLiveActivityTokenByID 删除失效 LA token 行。
func DeleteLiveActivityTokenByID(ctx context.Context, id uint64) error {
	if id == 0 {
		return nil
	}
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).Where("id", id).Delete()
	return err
}

// DeleteLiveActivityTokensByWxActivity 删除某用户全部 token_kind=activity 行。
// 业务：统一 LA end 后清 update token，避免脏行；MUST NOT 删除 push_to_start。
// Side Effects: DELETE push_live_activity_token。
func DeleteLiveActivityTokensByWxActivity(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return nil
	}
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("token_kind", LATokenKindActivity).
		Delete()
	return err
}
