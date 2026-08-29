package ucg

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 原力流水 reason（与 App 展示对齐）。
const (
	ForceReasonDebateSelfVote   = "debate_self_vote"   // 作者自投辩论 +1
	ForceReasonInviteAcquisition = "invite_acquisition" // 获客成功 +100
)

const ForceDeltaInviteAcquisition = 100

// EnsureForceSchema 创建原力计数与流水素表（ucg 本域，禁止再写 device.wx.force_value）。
func EnsureForceSchema(ctx context.Context) error {
	db := g.DB()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS ucg_user_force (
  wx_id       BIGINT NOT NULL,
  force_value INT    NOT NULL DEFAULT 0,
  updated_at  BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (wx_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS ucg_force_ledger (
  id         BIGINT       NOT NULL AUTO_INCREMENT,
  wx_id      BIGINT       NOT NULL,
  reason     VARCHAR(64)  NOT NULL DEFAULT '',
  delta      INT          NOT NULL DEFAULT 0,
  ref        VARCHAR(128) NOT NULL DEFAULT '',
  created_at BIGINT       NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_wx_time (wx_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, sql := range stmts {
		if _, err := db.Exec(ctx, sql); err != nil {
			return err
		}
	}
	return nil
}

// ForceLedgerEntry 积分流水行。
type ForceLedgerEntry struct {
	Id        int64  `json:"id"`
	Reason    string `json:"reason"`
	Delta     int    `json:"delta"`
	Ref       string `json:"ref,omitempty"`
	CreatedAt int64  `json:"createdAt"`
}

// GetForceValue 读取当前原力；无行则 0。
func GetForceValue(ctx context.Context, wxID int64) (int, error) {
	if wxID <= 0 {
		return 0, nil
	}
	var row struct {
		ForceValue int `json:"force_value"`
	}
	_ = g.DB().Model("ucg_user_force").Ctx(ctx).Where("wx_id", wxID).Scan(&row)
	return row.ForceValue, nil
}

// BatchForceValues 批量读原力。
func BatchForceValues(ctx context.Context, wxIDs []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(wxIDs))
	if len(wxIDs) == 0 {
		return out, nil
	}
	type rowT struct {
		WxId       int64 `json:"wx_id"`
		ForceValue int   `json:"force_value"`
	}
	var rows []rowT
	err := g.DB().Model("ucg_user_force").Ctx(ctx).WhereIn("wx_id", wxIDs).Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.WxId] = r.ForceValue
	}
	return out, nil
}

// AddForceDelta 原力增量并写流水（同事务）。
// Args: reason/delta/ref 业务标注；Side Effects: 写 ucg_user_force + ucg_force_ledger。
func AddForceDelta(ctx context.Context, wxID int64, reason string, delta int, ref string) error {
	if wxID <= 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	if delta == 0 {
		return nil
	}
	reason = strings.TrimSpace(reason)
	ref = strings.TrimSpace(ref)
	now := time.Now().Unix()
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var cur struct {
			ForceValue int `json:"force_value"`
		}
		err := tx.Model("ucg_user_force").Ctx(ctx).Where("wx_id", wxID).LockUpdate().Scan(&cur)
		if err != nil {
			return err
		}
		n, err := tx.Model("ucg_user_force").Ctx(ctx).Where("wx_id", wxID).Count()
		if err != nil {
			return err
		}
		if n == 0 {
			_, err = tx.Model("ucg_user_force").Ctx(ctx).Data(g.Map{
				"wx_id": wxID, "force_value": delta, "updated_at": now,
			}).Insert()
		} else {
			_, err = tx.Model("ucg_user_force").Ctx(ctx).Where("wx_id", wxID).Data(g.Map{
				"force_value": cur.ForceValue + delta,
				"updated_at":  now,
			}).Update()
		}
		if err != nil {
			return err
		}
		_, err = tx.Model("ucg_force_ledger").Ctx(ctx).Data(g.Map{
			"wx_id": wxID, "reason": reason, "delta": delta, "ref": ref, "created_at": now,
		}).Insert()
		return err
	})
}

// AddDebateSelfVoteForce 辩论自投 +1。
func AddDebateSelfVoteForce(ctx context.Context, wxID int64, postID int64) error {
	return AddForceDelta(ctx, wxID, ForceReasonDebateSelfVote, 1, g.NewVar(postID).String())
}

// AddInviteAcquisitionForce 获客成功 +100（供 cash 内部调用）。
func AddInviteAcquisitionForce(ctx context.Context, ownerWxID int64, ref string) error {
	return AddForceDelta(ctx, ownerWxID, ForceReasonInviteAcquisition, ForceDeltaInviteAcquisition, ref)
}

// ListForceLedger 按时间倒序分页流水。
func ListForceLedger(ctx context.Context, wxID int64, limit, offset int) ([]ForceLedgerEntry, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	type rowT struct {
		Id        int64  `json:"id"`
		Reason    string `json:"reason"`
		Delta     int    `json:"delta"`
		Ref       string `json:"ref"`
		CreatedAt int64  `json:"created_at"`
	}
	var rows []rowT
	err := g.DB().Model("ucg_force_ledger").Ctx(ctx).
		Where("wx_id", wxID).
		OrderDesc("created_at").OrderDesc("id").
		Limit(limit).Offset(offset).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]ForceLedgerEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, ForceLedgerEntry{
			Id: r.Id, Reason: r.Reason, Delta: r.Delta, Ref: r.Ref, CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}
