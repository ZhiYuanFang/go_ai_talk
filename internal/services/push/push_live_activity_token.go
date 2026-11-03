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

// Live Activity token 种类与业务 attributes 维度（与 Flutter Attributes 对齐，存小写 kind）。
const (
	LATokenKindPushToStart = "push_to_start"
	LATokenKindActivity    = "activity"

	LAAttributesImminent = "imminent"
	LAAttributesTiming   = "timing"
)

// entityLiveActivityToken LA 专用 token 行（与 push_device 分表）。
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

// RegisterLiveActivityToken 注册/更新 Push-to-Start 或 per-activity update token。
//
// 业务逻辑：
//  1. MUST NOT 写入 push_device（避免 uk_token 与 device alert token 互顶）；
//  2. 按 (wx, device_key, token_kind, attributes_kind, activity_id) upsert；
//  3. push_to_start 时 activity_id 强制空串；activity 时 activity_id 必填。
//
// Side Effects: 写 push_live_activity_token。
func RegisterLiveActivityToken(ctx context.Context, wxID int64, deviceKey, tokenKind, attributesKind, activityID, token string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	deviceKey = strings.TrimSpace(deviceKey)
	tokenKind = strings.TrimSpace(strings.ToLower(tokenKind))
	attributesKind = strings.TrimSpace(strings.ToLower(attributesKind))
	activityID = strings.TrimSpace(activityID)
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
	if attributesKind != LAAttributesImminent && attributesKind != LAAttributesTiming {
		return gerror.NewCode(gcode.CodeInvalidParameter, "attributesKind 必须是 imminent 或 timing")
	}
	if tokenKind == LATokenKindPushToStart {
		activityID = ""
	} else if activityID == "" || len(activityID) > 128 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "activity 类型须提供合法 activityId")
	}
	now := time.Now().Unix()
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).Data(g.Map{
		"wx_id":           wxID,
		"device_key":      deviceKey,
		"token_kind":      tokenKind,
		"attributes_kind": attributesKind,
		"activity_id":     activityID,
		"token":           token,
		"updated_at":      now,
	}).OnDuplicate("token", "updated_at").Save()
	if err != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "写入 Live Activity token 失败")
	}
	return nil
}

// ListLiveActivityTokensForWx 列出某用户某 kind/attributes 的 token（可多机）。
func ListLiveActivityTokensForWx(ctx context.Context, wxID int64, tokenKind, attributesKind string) ([]entityLiveActivityToken, error) {
	if wxID <= 0 {
		return nil, nil
	}
	tokenKind = strings.TrimSpace(strings.ToLower(tokenKind))
	attributesKind = strings.TrimSpace(strings.ToLower(attributesKind))
	var rows []entityLiveActivityToken
	err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("token_kind", tokenKind).
		Where("attributes_kind", attributesKind).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ListLiveActivityTokensByActivityID 按 wx + activityId 取 update/end 用 token。
func ListLiveActivityTokensByActivityID(ctx context.Context, wxID int64, activityID string) ([]entityLiveActivityToken, error) {
	if wxID <= 0 {
		return nil, nil
	}
	activityID = strings.TrimSpace(activityID)
	if activityID == "" {
		return nil, nil
	}
	var rows []entityLiveActivityToken
	err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("token_kind", LATokenKindActivity).
		Where("activity_id", activityID).
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
