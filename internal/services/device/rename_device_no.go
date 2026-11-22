package device

import (
	"context"
	"fmt"
	"strings"

	cashclient "hello/internal/clients/cash"
	gwclient "hello/internal/clients/gatewayapp"
	histclient "hello/internal/clients/history"
	"hello/internal/dao"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// RenameDeviceNoResult 改名编排结果（分步可观测）。
type RenameDeviceNoResult struct {
	OldDeviceNo string  `json:"oldDeviceNo"`
	NewDeviceNo string  `json:"newDeviceNo"`
	WxIds       []int64 `json:"wxIds"`
	// Steps 记录已完成步骤，便于半失败后重试对照。
	Steps []string `json:"steps"`
}

// RenameDeviceNo 宝宝号原地改名编排（Hub Admin）。
// 顺序：本域表 → history migrate → cash migrate → 吊销会话。
// 幂等：若 user 已是 new 且 old 无注册，跳过本域写，继续补下游。
func RenameDeviceNo(ctx context.Context, oldDeviceNo, newDeviceNo string, generateNew bool) (*RenameDeviceNoResult, error) {
	old := strings.TrimSpace(oldDeviceNo)
	neu := strings.TrimSpace(newDeviceNo)
	if old == "" {
		return nil, fmt.Errorf("oldDeviceNo 不能为空")
	}
	if generateNew {
		dn, err := generateUniqueDeviceNo(ctx)
		if err != nil {
			return nil, err
		}
		neu = dn
	}
	if neu == "" {
		return nil, fmt.Errorf("newDeviceNo 不能为空")
	}
	if old == neu {
		return nil, fmt.Errorf("新旧 deviceNo 不能相同")
	}

	out := &RenameDeviceNoResult{OldDeviceNo: old, NewDeviceNo: neu, Steps: nil}

	oldExists := false
	if err := DeviceAdmin().EnsureRegistered(ctx, old); err == nil {
		oldExists = true
	} else if err != ErrDeviceNotRegistered {
		return nil, err
	}
	newExists := false
	if err := DeviceAdmin().EnsureRegistered(ctx, neu); err == nil {
		newExists = true
	} else if err != ErrDeviceNotRegistered {
		return nil, err
	}

	if !oldExists && !newExists {
		return nil, fmt.Errorf("旧号未注册且新号亦不存在，无法改名或重试")
	}
	if oldExists && newExists {
		return nil, fmt.Errorf("新号已被占用")
	}

	// 列出绑定 wx：改名前用 old；幂等重试用 new。
	listNo := old
	if !oldExists {
		listNo = neu
	}
	wxIDs, _, err := ListWxIDsByDeviceNo(ctx, listNo)
	if err != nil {
		return nil, fmt.Errorf("列出绑定账号失败: %w", err)
	}
	out.WxIds = wxIDs

	// 先探再改：下游不可达时不得先改本域（首次路径）。
	if err := probeRenameDownstream(ctx); err != nil {
		return out, err
	}
	out.Steps = append(out.Steps, "probe_ok")

	if oldExists {
		if err := renameDeviceNoLocal(ctx, old, neu, wxIDs); err != nil {
			return out, fmt.Errorf("本域迁移失败: %w", err)
		}
		out.Steps = append(out.Steps, "device_local")
	} else {
		out.Steps = append(out.Steps, "device_local_skip_idempotent")
	}

	if err := histclient.RemoteMigrateDeviceNo(ctx, old, neu); err != nil {
		return out, fmt.Errorf("history 迁移失败（可重试）: %w", err)
	}
	out.Steps = append(out.Steps, "history")

	if err := cashclient.RemoteMigrateDeviceNo(ctx, old, neu); err != nil {
		return out, fmt.Errorf("cash 迁移失败（可重试）: %w", err)
	}
	out.Steps = append(out.Steps, "cash")

	if err := gwclient.RevokeSessions(ctx, wxIDs, old); err != nil {
		return out, fmt.Errorf("会话吊销失败（可重试）: %w", err)
	}
	out.Steps = append(out.Steps, "revoke_sessions")

	glog.Infof(ctx, "[device] RenameDeviceNo ok oldLen=%d newLen=%d wxCount=%d steps=%v",
		len(old), len(neu), len(wxIDs), out.Steps)
	return out, nil
}

// probeRenameDownstream 探测 history / cash / gateway 内部契约可达（dryRun，不改写）。
func probeRenameDownstream(ctx context.Context) error {
	if err := histclient.RemoteProbeMigrateDeviceNo(ctx); err != nil {
		return fmt.Errorf("改名预检失败（history）: %w", err)
	}
	if err := cashclient.RemoteProbeMigrateDeviceNo(ctx); err != nil {
		return fmt.Errorf("改名预检失败（cash）: %w", err)
	}
	if err := gwclient.ProbeRevokeSessions(ctx); err != nil {
		return fmt.Errorf("改名预检失败（gateway）: %w", err)
	}
	return nil
}

// renameDeviceNoLocal 更新 device 库内宝宝分区键与绑定，并清理旧 Redis。
func renameDeviceNoLocal(ctx context.Context, old, neu string, wxIDs []int64) error {
	db := g.DB(dao.User.Group())
	err := db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		res, err := tx.Model(dao.User.Table()).Ctx(ctx).
			Data(dao.User.Columns().DeviceNo, neu).
			Where(dao.User.Columns().DeviceNo, old).
			Update()
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("user 行未更新（旧号可能已不存在）")
		}
		if _, err := tx.Model(dao.Wx.Table()).Ctx(ctx).
			Data(dao.Wx.Columns().DeviceNo, neu).
			Where(dao.Wx.Columns().DeviceNo, old).
			Update(); err != nil {
			return fmt.Errorf("wx: %w", err)
		}
		if _, err := tx.Model(dao.AppointmentNext.Table()).Ctx(ctx).
			Data("device_no", neu).
			Where("device_no", old).
			Update(); err != nil {
			return fmt.Errorf("appointment_next: %w", err)
		}
		if _, err := tx.Model(dao.XiaozhiMcpBinding.Table()).Ctx(ctx).
			Data("device_no", neu).
			Where("device_no", old).
			Update(); err != nil {
			return fmt.Errorf("xiaozhi_mcp_binding: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	delDeviceRedisBestEffort(ctx, old, wxIDs)
	return nil
}

func delDeviceRedisBestEffort(ctx context.Context, oldDeviceNo string, wxIDs []int64) {
	c := cachekit.Default()
	if key, err := cachekit.UserProfileKey(oldDeviceNo); err == nil && key != "" {
		_ = c.Del(ctx, key)
	}
	if key, err := cachekit.UserProfileVersionKey(oldDeviceNo); err == nil && key != "" {
		_ = c.Del(ctx, key)
	}
	for _, id := range wxIDs {
		if id <= 0 {
			continue
		}
		_ = c.Del(ctx, cachekit.WxIDToDeviceKey(id))
	}
}
