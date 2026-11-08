// Live Activity update 送达回执：临时 ack + 可配置短超时；超时未 ack 则删 activity 并 push_to_start 重启。
// AfterFunc 仅为单次超时，非扫表 ticker。
package push

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	liveActivityAckTimeoutEnv     = "LIVE_ACTIVITY_ACK_TIMEOUT_MS"
	liveActivityAckTimeoutDefault = 3000
	liveActivityAckTimeoutMaxMS   = 60000
	liveActivityAckContentKey     = "ack"
)

// LiveActivityAckTimeout 返回 update 等待客户端 ack 的超时。
// 业务：LIVE_ACTIVITY_ACK_TIMEOUT_MS；空/非法回退 3000；上限 60000。
func LiveActivityAckTimeout() time.Duration {
	raw := strings.TrimSpace(os.Getenv(liveActivityAckTimeoutEnv))
	if raw == "" {
		return time.Duration(liveActivityAckTimeoutDefault) * time.Millisecond
	}
	ms, err := strconv.Atoi(raw)
	if err != nil || ms <= 0 {
		return time.Duration(liveActivityAckTimeoutDefault) * time.Millisecond
	}
	if ms > liveActivityAckTimeoutMaxMS {
		ms = liveActivityAckTimeoutMaxMS
	}
	return time.Duration(ms) * time.Millisecond
}

// newLiveActivityAck 生成不透明临时序号（ContentState.ack / 行 last_sent_ack）。
func newLiveActivityAck() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
}

// cloneContentStateWithAck 复制业务 ContentState 并写入 ack（避免污染调用方 map / 多机互串）。
func cloneContentStateWithAck(src map[string]interface{}, ack string) map[string]interface{} {
	out := make(map[string]interface{}, len(src)+1)
	for k, v := range src {
		out[k] = v
	}
	out[liveActivityAckContentKey] = ack
	return out
}

// MarkLiveActivityAckSent update send_ok 后写入 last_sent_ack / ack_sent_at。
// Side Effects: UPDATE push_live_activity_token。
func MarkLiveActivityAckSent(ctx context.Context, id uint64, ack string) error {
	if id == 0 || strings.TrimSpace(ack) == "" {
		return nil
	}
	now := time.Now().Unix()
	_, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("id", id).
		Data(g.Map{
			"last_sent_ack": ack,
			"ack_sent_at":   now,
			"updated_at":    now,
		}).Update()
	return err
}

// AckLiveActivityUpdate 客户端回传 ack：仅当等于当前 last_sent_ack 时写入 last_acked_ack。
//
// 业务：对齐 register 免鉴权身份（body wxId）；乱序旧 ack MUST NOT 回写；行已删则 no-op。
// Args: wxID、deviceKey、ack。
// Returns: 错误（参数/DB）；未匹配 sent 不报错。
// Side Effects: 可能 UPDATE last_acked_ack。
func AckLiveActivityUpdate(ctx context.Context, wxID int64, deviceKey, ack string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	deviceKey = strings.TrimSpace(deviceKey)
	ack = strings.TrimSpace(ack)
	if deviceKey == "" || len(deviceKey) > 64 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "deviceKey 无效")
	}
	if ack == "" || len(ack) > 64 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "ack 无效")
	}
	row, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("wx_id", wxID).
		Where("device_key", deviceKey).
		Where("token_kind", LATokenKindActivity).
		One()
	if err != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, err, "查询 Live Activity token 失败")
	}
	if row.IsEmpty() {
		glog.Infof(ctx, "[push-la] la_ack ignore reason=no_activity_row wxId=%d deviceKeyLen=%d", wxID, len(deviceKey))
		return nil
	}
	var tok entityLiveActivityToken
	if scanErr := row.Struct(&tok); scanErr != nil {
		return gerror.WrapCode(gcode.CodeInternalError, scanErr, "解析 Live Activity token 失败")
	}
	if strings.TrimSpace(tok.LastSentAck) != ack {
		glog.Infof(ctx, "[push-la] la_ack ignore reason=mismatch wxId=%d tokenId=%d", wxID, tok.ID)
		return nil
	}
	now := time.Now().Unix()
	if _, updErr := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).
		Where("id", tok.ID).
		Where("last_sent_ack", ack). // 防并发被新一轮 update 覆盖后仍写旧 ack
		Data(g.Map{
			"last_acked_ack": ack,
			"updated_at":     now,
		}).Update(); updErr != nil {
		return gerror.WrapCode(gcode.CodeDbOperationError, updErr, "写入 Live Activity ack 失败")
	}
	glog.Infof(ctx, "[push-la] la_ack_ok wxId=%d tokenId=%d", wxID, tok.ID)
	return nil
}

// liveActivityAckTimedOut 行上已发出未确认且超过超时（供 AfterFunc 与下次 update 补刀）。
func liveActivityAckTimedOut(tok entityLiveActivityToken, timeout time.Duration) bool {
	sent := strings.TrimSpace(tok.LastSentAck)
	if sent == "" {
		return false
	}
	if sent == strings.TrimSpace(tok.LastAckedAck) {
		return false
	}
	if tok.AckSentAt <= 0 {
		return false
	}
	deadline := time.Unix(tok.AckSentAt, 0).Add(timeout)
	return !time.Now().Before(deadline)
}

// scheduleLiveActivityAckTimeout update send_ok 后单次 AfterFunc；到期未 ack 则失效重启。
// 业务：非扫表 ticker；行已删 / 已 ack / sent 已换新则跳过。
// Side Effects: 可能 DELETE activity + APNs start。
func scheduleLiveActivityAckTimeout(op LiveActivityOp, tok entityLiveActivityToken, sentAck string) {
	timeout := LiveActivityAckTimeout()
	tokenID := tok.ID
	wxID := op.WxID
	time.AfterFunc(timeout, func() {
		ctx := context.Background()
		row, err := g.DB().Model(pushLiveActivityTokenTable).Ctx(ctx).Where("id", tokenID).One()
		if err != nil {
			g.Log().Warningf(ctx, "[push-la] la_ack_timeout check_failed tokenId=%d err=%v", tokenID, err)
			return
		}
		if row.IsEmpty() {
			return
		}
		var cur entityLiveActivityToken
		if scanErr := row.Struct(&cur); scanErr != nil {
			g.Log().Warningf(ctx, "[push-la] la_ack_timeout scan_failed tokenId=%d err=%v", tokenID, scanErr)
			return
		}
		if strings.TrimSpace(cur.LastSentAck) != strings.TrimSpace(sentAck) {
			return
		}
		if strings.TrimSpace(cur.LastAckedAck) == strings.TrimSpace(sentAck) {
			return
		}
		g.Log().Warningf(ctx, "[push-la] la_ack_timeout_restart wxId=%d tokenId=%d deviceKey=%s timeoutMs=%d",
			wxID, tokenID, cur.DeviceKey, timeout.Milliseconds())
		if delErr := DeleteLiveActivityTokenByID(ctx, tokenID); delErr != nil {
			g.Log().Warningf(ctx, "[push-la] la_ack_timeout delete_failed id=%d err=%v", tokenID, delErr)
			return
		}
		restartLiveActivityViaPushToStart(ctx, op, cur)
	})
}

// failoverStaleActivityIfAckTimedOut 下次 update 前 DB 补刀：已超时未 ack 则删行并 start，返回 true 表示本行勿再 update。
func failoverStaleActivityIfAckTimedOut(ctx context.Context, op LiveActivityOp, tok entityLiveActivityToken) bool {
	if !liveActivityAckTimedOut(tok, LiveActivityAckTimeout()) {
		return false
	}
	g.Log().Warningf(ctx, "[push-la] la_ack_stale_before_update wxId=%d tokenId=%d deviceKey=%s",
		op.WxID, tok.ID, tok.DeviceKey)
	if delErr := DeleteLiveActivityTokenByID(ctx, tok.ID); delErr != nil {
		g.Log().Warningf(ctx, "[push-la] la_ack_stale delete_failed id=%d err=%v", tok.ID, delErr)
		return true
	}
	restartLiveActivityViaPushToStart(ctx, op, tok)
	return true
}
