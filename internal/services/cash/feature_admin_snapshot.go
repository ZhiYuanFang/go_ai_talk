package cash

// Admin 开通快照（方案 A：当前权益/条数权威表，非事件流水）。

import (
	"context"
	"strings"
	"time"

	ucgclient "hello/internal/clients/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// FeatureActivationSnapshotItem 单条当前开通快照。
type FeatureActivationSnapshotItem struct {
	SubjectType       string `json:"subjectType"` // device | user | count
	DeviceNo          string `json:"deviceNo,omitempty"`
	WxId              int64  `json:"wxId,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	UnlockMethod      string `json:"unlockMethod,omitempty"`
	PermanentDelta    int    `json:"permanentDelta,omitempty"`
	ExpiresAt         int64  `json:"expiresAt"` // 0=永久；预测条数类可为 0 且 Kind=count
	Active            bool   `json:"active"`
	RemainingSeconds  int64  `json:"remainingSeconds,omitempty"`
	UpdatedAt         int64  `json:"updatedAt,omitempty"`
	Kind              string `json:"kind"` // entitlement | allowed_count
}

// FeatureActivationSnapshotPage 分页快照。
type FeatureActivationSnapshotPage struct {
	FeatureId   string                          `json:"featureId"`
	RuleSummary string                          `json:"ruleSummary"`
	Note        string                          `json:"note"`
	Total       int                             `json:"total"`
	List        []FeatureActivationSnapshotItem `json:"list"`
}

const activationSnapshotNote = "本页为当前开通快照（非完整历史事件）；不含 VIP 旁路覆盖。"

// AdminListFeatureActivationSnapshot 按功能列出当前开通主体快照。
//
// 业务：device 读 feature_entitlement；user 读 feature_user_entitlement；预测读 feature_allowed_count。
// Args: featureID；limit/offset 分页（limit 默认 50，最大 200）。
// Returns: 快照页；未知功能报错。
func AdminListFeatureActivationSnapshot(ctx context.Context, featureID string, limit, offset int) (*FeatureActivationSnapshotPage, error) {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "featureId 不能为空")
	}
	r, err := g.DB().Model("feature_def").Ctx(ctx).Where("feature_id", featureID).One()
	if err != nil {
		return nil, err
	}
	if r.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "功能不存在")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	now := time.Now().Unix()
	page := &FeatureActivationSnapshotPage{
		FeatureId:   featureID,
		RuleSummary: FeatureRuleSummary(featureID),
		Note:        activationSnapshotNote,
		List:        make([]FeatureActivationSnapshotItem, 0),
	}

	if featureID == FeatureIDPredictionUnlock {
		total, err := g.DB().Model("feature_allowed_count").Ctx(ctx).Count()
		if err != nil {
			return nil, err
		}
		page.Total = total
		type rowT struct {
			DeviceNo     string `json:"device_no"`
			AllowedCount int    `json:"allowed_count"`
			UpdatedAt    int64  `json:"updated_at"`
		}
		var rows []rowT
		err = g.DB().Model("feature_allowed_count").Ctx(ctx).
			Fields("device_no,allowed_count,updated_at").
			OrderDesc("updated_at").OrderDesc("device_no").
			Limit(limit).Offset(offset).Scan(&rows)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			page.List = append(page.List, FeatureActivationSnapshotItem{
				SubjectType:    ActivationSubjectDevice,
				DeviceNo:       row.DeviceNo,
				PermanentDelta: row.AllowedCount,
				ExpiresAt:      0,
				Active:         row.AllowedCount > 0,
				UpdatedAt:      row.UpdatedAt,
				Kind:           "allowed_count",
			})
		}
		return page, nil
	}

	subj := NormalizeActivationSubject(r["activation_subject"].String())
	if subj == ActivationSubjectUser {
		total, err := g.DB().Model("feature_user_entitlement").Ctx(ctx).Where("feature_id", featureID).Count()
		if err != nil {
			return nil, err
		}
		page.Total = total
		type rowT struct {
			WxId         int64  `json:"wx_id"`
			UnlockMethod string `json:"unlock_method"`
			ExpiresAt    int64  `json:"expires_at"`
			UpdatedAt    int64  `json:"updated_at"`
		}
		var rows []rowT
		err = g.DB().Model("feature_user_entitlement").Ctx(ctx).
			Where("feature_id", featureID).
			Fields("wx_id,unlock_method,expires_at,updated_at").
			OrderDesc("updated_at").OrderDesc("wx_id").
			Limit(limit).Offset(offset).Scan(&rows)
		if err != nil {
			return nil, err
		}
		ids := make([]int64, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.WxId)
		}
		nicks, _ := ucgclient.FetchUcgNicknames(ctx, ids)
		for _, row := range rows {
			active, remain := entitlementActiveRemain(now, row.ExpiresAt)
			page.List = append(page.List, FeatureActivationSnapshotItem{
				SubjectType:      ActivationSubjectUser,
				WxId:             row.WxId,
				Nickname:         nicks[row.WxId],
				UnlockMethod:     row.UnlockMethod,
				ExpiresAt:        row.ExpiresAt,
				Active:           active,
				RemainingSeconds: remain,
				UpdatedAt:        row.UpdatedAt,
				Kind:             "entitlement",
			})
		}
		return page, nil
	}

	total, err := g.DB().Model("feature_entitlement").Ctx(ctx).Where("feature_id", featureID).Count()
	if err != nil {
		return nil, err
	}
	page.Total = total
	type rowT struct {
		DeviceNo     string `json:"device_no"`
		UnlockMethod string `json:"unlock_method"`
		ExpiresAt    int64  `json:"expires_at"`
		UpdatedAt    int64  `json:"updated_at"`
	}
	var rows []rowT
	err = g.DB().Model("feature_entitlement").Ctx(ctx).
		Where("feature_id", featureID).
		Fields("device_no,unlock_method,expires_at,updated_at").
		OrderDesc("updated_at").OrderDesc("device_no").
		Limit(limit).Offset(offset).Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		active, remain := entitlementActiveRemain(now, row.ExpiresAt)
		page.List = append(page.List, FeatureActivationSnapshotItem{
			SubjectType:      ActivationSubjectDevice,
			DeviceNo:         row.DeviceNo,
			UnlockMethod:     row.UnlockMethod,
			ExpiresAt:        row.ExpiresAt,
			Active:           active,
			RemainingSeconds: remain,
			UpdatedAt:        row.UpdatedAt,
			Kind:             "entitlement",
		})
	}
	return page, nil
}

// entitlementActiveRemain 根据 expires_at 计算是否有效与剩余秒（永久 remain=0 且 active=true）。
func entitlementActiveRemain(now, expiresAt int64) (active bool, remain int64) {
	if expiresAt == 0 {
		return true, 0
	}
	if expiresAt > now {
		return true, expiresAt - now
	}
	return false, 0
}
