package cash

// xiaozhi_mcp_admin_list.go：Hub 小智 MCP 开通人员快照列表（Admin）。
//
// 业务：分页读 xiaozhi_mcp_entitlement 全表（含已撤销/试用过期），经 ucg 批量补昵称；
// 对齐开通功能管理 activations 快照交互，供区 C 展示与行内手工授撤销入口判断。

import (
	"context"
	"time"

	ucgclient "hello/internal/clients/ucg"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// XiaozhiMcpEntitlementListItem 开通人员快照行。
type XiaozhiMcpEntitlementListItem struct {
	WxId             int64  `json:"wxId"`
	Nickname         string `json:"nickname,omitempty"`
	UnlockMethod     string `json:"unlockMethod"`
	ChannelRef       string `json:"channelRef,omitempty"`
	UnlockedAt       int64  `json:"unlockedAt"`
	ExpiresAt        int64  `json:"expiresAt"`
	Active           bool   `json:"active"`
	RemainingSeconds int64  `json:"remainingSeconds,omitempty"`
	Status           int    `json:"status"`
	RevokedAt        int64  `json:"revokedAt,omitempty"`
	UpdatedAt        int64  `json:"updatedAt"`
}

// XiaozhiMcpEntitlementListPage 分页快照。
type XiaozhiMcpEntitlementListPage struct {
	Note  string                          `json:"note"`
	Total int                             `json:"total"`
	List  []XiaozhiMcpEntitlementListItem `json:"list"`
}

const xiaozhiMcpEntitlementListNote = "当前权益快照（非完整历史事件）；含已撤销与试用过期行。"

// AdminListXiaozhiMcpEntitlements 管理端分页列出小智 MCP 开通人员。
//
// 业务：全表按 updated_at/wx_id 倒序；active=status 有效且（永久或试用未过期）；昵称失败降级空串。
// Args: limit 默认 50、最大 200；offset 默认 0。
// Returns: 快照页；DB 错误原样返回。
func AdminListXiaozhiMcpEntitlements(ctx context.Context, limit, offset int) (*XiaozhiMcpEntitlementListPage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	page := &XiaozhiMcpEntitlementListPage{
		Note: xiaozhiMcpEntitlementListNote,
		List: make([]XiaozhiMcpEntitlementListItem, 0),
	}
	total, err := g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).Count()
	if err != nil {
		return nil, err
	}
	page.Total = total

	type rowT struct {
		WxId         int64  `json:"wx_id"`
		Status       int    `json:"status"`
		UnlockMethod string `json:"unlock_method"`
		ChannelRef   string `json:"channel_ref"`
		UnlockedAt   int64  `json:"unlocked_at"`
		ExpiresAt    int64  `json:"expires_at"`
		RevokedAt    int64  `json:"revoked_at"`
		UpdatedAt    int64  `json:"updated_at"`
	}
	var rows []rowT
	// 列表 Scan 合法；空表得到空 slice。
	err = g.DB().Model("xiaozhi_mcp_entitlement").Ctx(ctx).
		Fields("wx_id,status,unlock_method,channel_ref,unlocked_at,expires_at,revoked_at,updated_at").
		OrderDesc("updated_at").OrderDesc("wx_id").
		Limit(limit).Offset(offset).Scan(&rows)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.WxId)
	}
	nicks, nickErr := ucgclient.FetchUcgNicknames(ctx, ids)
	if nickErr != nil {
		glog.Warningf(ctx, "[cash] xiaozhi-mcp entitlements nickname fetch failed err=%v", nickErr)
		nicks = map[int64]string{}
	}

	now := time.Now().Unix()
	for _, row := range rows {
		active, remain := xiaozhiMcpEntitlementActiveRemain(now, row.Status, row.ExpiresAt)
		page.List = append(page.List, XiaozhiMcpEntitlementListItem{
			WxId:             row.WxId,
			Nickname:         nicks[row.WxId],
			UnlockMethod:     row.UnlockMethod,
			ChannelRef:       row.ChannelRef,
			UnlockedAt:       row.UnlockedAt,
			ExpiresAt:        row.ExpiresAt,
			Active:           active,
			RemainingSeconds: remain,
			Status:           row.Status,
			RevokedAt:        row.RevokedAt,
			UpdatedAt:        row.UpdatedAt,
		})
	}
	return page, nil
}

// xiaozhiMcpEntitlementActiveRemain 计算列表行是否仍有效及试用剩余秒。
//
// 业务：status 非有效 → 无效；expires_at=0 → 永久有效 remain=0；未过期试用 → 有效并返回剩余秒；否则无效。
func xiaozhiMcpEntitlementActiveRemain(now int64, status int, expiresAt int64) (active bool, remain int64) {
	if status != XiaozhiMcpEntitlementActive {
		return false, 0
	}
	if expiresAt == 0 {
		return true, 0
	}
	if expiresAt > now {
		return true, expiresAt - now
	}
	return false, 0
}
