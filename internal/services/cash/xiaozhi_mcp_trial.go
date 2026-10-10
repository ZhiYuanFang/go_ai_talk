package cash

// xiaozhi_mcp_trial.go：小智 MCP 每账号一次试用（24h）与 Add 前门禁 EnsureAccessForAdd。

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// IsXiaozhiMcpTrialUnused 是否尚未成功领取试用（无行或 unused）。
func IsXiaozhiMcpTrialUnused(ctx context.Context, wxID int64) (bool, error) {
	if wxID <= 0 {
		return false, nil
	}
	one, err := g.DB().Model("xiaozhi_mcp_trial").Ctx(ctx).
		Fields("status").Where("wx_id", wxID).One()
	if err != nil {
		return false, err
	}
	if one.IsEmpty() {
		return true, nil
	}
	return strings.TrimSpace(one["status"].String()) == XiaozhiMcpTrialStatusUnused, nil
}

// ClaimXiaozhiMcpTrial 领取 24h 试用：unused→used + 限时权益。
// 已有效权益时幂等跳过（不耗试用）。已 used 且无权益则拒绝。
func ClaimXiaozhiMcpTrial(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	active, err := HasActiveXiaozhiMcpEntitlement(ctx, wxID)
	if err != nil {
		return err
	}
	if active {
		return nil
	}
	unused, err := IsXiaozhiMcpTrialUnused(ctx, wxID)
	if err != nil {
		return err
	}
	if !unused {
		return gerror.NewCode(gcode.CodeInvalidOperation, "试用已用尽，请付费开通小智 MCP")
	}
	now := time.Now().Unix()
	// 确保有 unused 行。
	_, _ = g.DB().Exec(ctx, `
INSERT IGNORE INTO xiaozhi_mcp_trial (wx_id, status, used_at, updated_at)
VALUES (?, ?, 0, ?)`, wxID, XiaozhiMcpTrialStatusUnused, now)

	res, err := g.DB().Model("xiaozhi_mcp_trial").Ctx(ctx).
		Where("wx_id", wxID).Where("status", XiaozhiMcpTrialStatusUnused).
		Data(g.Map{
			"status":     XiaozhiMcpTrialStatusUsed,
			"used_at":    now,
			"updated_at": now,
		}).Update()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return gerror.NewCode(gcode.CodeInvalidOperation, "试用已用尽，请付费开通小智 MCP")
	}
	exp := now + int64(TrialDurationHours)*3600
	if err = grantXiaozhiMcpTimed(ctx, wxID, XiaozhiMcpUnlockTrial, "trial", exp); err != nil {
		_ = revertXiaozhiMcpTrialUnused(ctx, wxID)
		return err
	}
	glog.Infof(ctx, "[cash] xiaozhi-mcp trial claimed wxId=%d expiresAt=%d hours=%d", wxID, exp, TrialDurationHours)
	return nil
}

func revertXiaozhiMcpTrialUnused(ctx context.Context, wxID int64) error {
	now := time.Now().Unix()
	_, err := g.DB().Model("xiaozhi_mcp_trial").Ctx(ctx).
		Where("wx_id", wxID).Where("status", XiaozhiMcpTrialStatusUsed).
		Data(g.Map{
			"status":     XiaozhiMcpTrialStatusUnused,
			"used_at":    0,
			"updated_at": now,
		}).Update()
	return err
}

// grantXiaozhiMcpTimed 写入限时权益（试用）；不覆盖已有永久有效行。
func grantXiaozhiMcpTimed(ctx context.Context, wxID int64, unlockMethod, channelRef string, expiresAt int64) error {
	if wxID <= 0 || expiresAt <= time.Now().Unix() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "试用截止无效")
	}
	now := time.Now().Unix()
	one, err := g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Where("wx_id", wxID).One()
	if err != nil {
		return err
	}
	// 已有永久有效：勿降级。
	if !one.IsEmpty() && one["status"].Int() == XiaozhiMcpEntitlementActive && one["expires_at"].Int64() == 0 {
		return nil
	}
	data := g.Map{
		"status":        XiaozhiMcpEntitlementActive,
		"unlock_method": unlockMethod,
		"channel_ref":   channelRef,
		"expires_at":    expiresAt,
		"unlocked_at":   now,
		"revoked_at":    0,
		"updated_at":    now,
	}
	if one.IsEmpty() {
		data["wx_id"] = wxID
		_, err = g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Where("wx_id", wxID).Data(data).Update()
	return err
}

// EnsureXiaozhiMcpAccessForAdd Add 前确保有效权益：已有效放行；可试用则 claim；否则拒绝。
func EnsureXiaozhiMcpAccessForAdd(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	active, err := HasActiveXiaozhiMcpEntitlement(ctx, wxID)
	if err != nil {
		return err
	}
	if active {
		return nil
	}
	return ClaimXiaozhiMcpTrial(ctx, wxID)
}
