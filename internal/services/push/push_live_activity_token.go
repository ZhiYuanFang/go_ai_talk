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
// 唯一键为 (wx, device_key, token_kind)；不再含 attributes_kind / activity_id / ack 回执列。
type entityLiveActivityToken struct {
	ID        uint64
	WxID      int64
	DeviceKey string
	TokenKind string
	Token     string
	UpdatedAt int64
}

// RegisterLiveActivityToken 注册/更新 Push-to-Start 或 per-user activity update token。
//
// 业务逻辑：
//  1. MUST NOT 写入 push_device（避免 uk_token 与 device alert token 互顶）；
//  2. 按 (wx, device_key, token_kind) upsert；不再要求/写入 attributesKind/activityId。
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
	data := g.Map{
		"wx_id":      wxID,
		"device_key": deviceKey,
		"token_kind": tokenKind,
		"token":      token,
		"updated_at": now,
	}
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).Data(data).
		OnDuplicate("token", "updated_at").Save()
	if err != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "写入 Live Activity token 失败")
	}
	return nil
}

// ClearLiveActivityActivityToken 按 (wx, deviceKey) 删除 token_kind=activity 行。
//
// 业务：App 打开发现本机无 LA 时由鉴权接口调用；无行幂等成功；MUST NOT 删 push_to_start；MUST NOT 同拍 start。
// Side Effects: 可能 DELETE push_live_activity_token。
func ClearLiveActivityActivityToken(ctx context.Context, wxID int64, deviceKey string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	deviceKey = strings.TrimSpace(deviceKey)
	if deviceKey == "" || len(deviceKey) > 64 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceKey 无效")
	}
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("device_key", deviceKey).
		Where("token_kind", LATokenKindActivity).
		Delete()
	if err != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "清除 Live Activity activity token 失败")
	}
	g.Log().Warningf(ctx, "[push-la] activity_token_cleared wxId=%d deviceKey=%s", wxID, deviceKey)
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

// FindLiveActivityTokenByWxDeviceKind 按 (wx, device_key, token_kind) 取单行。
//
// 业务：按机精确查找 LA token；无行返回 ok=false。
// Args: wxID、deviceKey、tokenKind（通常为 push_to_start）。
// Returns: 行、是否找到、错误。
func FindLiveActivityTokenByWxDeviceKind(ctx context.Context, wxID int64, deviceKey, tokenKind string) (entityLiveActivityToken, bool, error) {
	var zero entityLiveActivityToken
	if wxID <= 0 {
		return zero, false, nil
	}
	deviceKey = strings.TrimSpace(deviceKey)
	tokenKind = strings.TrimSpace(strings.ToLower(tokenKind))
	if deviceKey == "" || tokenKind == "" {
		return zero, false, nil
	}
	row, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("device_key", deviceKey).
		Where("token_kind", tokenKind).
		One()
	if err != nil {
		return zero, false, err
	}
	if row.IsEmpty() {
		return zero, false, nil
	}
	var out entityLiveActivityToken
	if scanErr := row.Struct(&out); scanErr != nil {
		return zero, false, scanErr
	}
	return out, true, nil
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
