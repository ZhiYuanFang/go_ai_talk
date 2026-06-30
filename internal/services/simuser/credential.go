package simuser

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// WxCredential 模拟用户注册凭据（明文密码仅 admin 与任务登录使用）。
type WxCredential struct {
	WxId          int64  `json:"wxId"`
	Account       string `json:"account"`
	PasswordPlain string `json:"passwordPlain"`
	CreatedAt     int64  `json:"createdAt"`
}

// InsertWxCredential T1 注册成功后写入凭据。
func InsertWxCredential(ctx context.Context, wxID int64, account, passwordPlain string) error {
	if wxID <= 0 {
		return nil
	}
	now := time.Now().Unix()
	_, err := g.DB().Model("sim_wx_credential").Ctx(ctx).Data(g.Map{
		"wx_id": wxID, "account": account,
		"password_plain": passwordPlain, "created_at": now,
	}).Save()
	return err
}

// GetWxCredentialsByWxIDs 批量读取凭据；缺失 wxId 不出现在 map。
func GetWxCredentialsByWxIDs(ctx context.Context, wxIDs []int64) (map[int64]WxCredential, error) {
	out := make(map[int64]WxCredential)
	if len(wxIDs) == 0 {
		return out, nil
	}
	rows, err := g.DB().Model("sim_wx_credential").Ctx(ctx).
		WhereIn("wx_id", wxIDs).All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		id := row["wx_id"].Int64()
		out[id] = WxCredential{
			WxId:          id,
			Account:       row["account"].String(),
			PasswordPlain: row["password_plain"].String(),
			CreatedAt:     row["created_at"].Int64(),
		}
	}
	return out, nil
}

// GetWxCredentialByWxID 读取单条凭据。
func GetWxCredentialByWxID(ctx context.Context, wxID int64) (WxCredential, bool, error) {
	m, err := GetWxCredentialsByWxIDs(ctx, []int64{wxID})
	if err != nil {
		return WxCredential{}, false, err
	}
	c, ok := m[wxID]
	return c, ok, nil
}

// DeleteWxCredentialByWxID 注销后删除凭据行。
func DeleteWxCredentialByWxID(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return nil
	}
	_, err := g.DB().Model("sim_wx_credential").Ctx(ctx).Where("wx_id", wxID).Delete()
	return err
}

// SkipPendingVideoJobsForWx 注销时将未完成视频 job 标为 skipped。
func SkipPendingVideoJobsForWx(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return nil
	}
	_, err := g.DB().Model("sim_video_job").Ctx(ctx).
		Where("wx_id", wxID).
		WhereIn("status", []string{"pending", "processing"}).
		Data(g.Map{"status": "skipped", "updated_at": time.Now().Unix()}).
		Update()
	return err
}

const legacySimDefaultPassword = "123456"

// ResolveSimLoginPassword 按 wxId 取登录密码；无 credential 时 fallback 历史默认密码。
func ResolveSimLoginPassword(ctx context.Context, wxID int64, fallback string) (string, error) {
	if cred, ok, err := GetWxCredentialByWxID(ctx, wxID); err != nil {
		return "", err
	} else if ok && cred.PasswordPlain != "" {
		return cred.PasswordPlain, nil
	}
	if fallback = trimFallbackPassword(fallback); fallback != "" {
		return fallback, nil
	}
	return legacySimDefaultPassword, nil
}

func trimFallbackPassword(fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback == "" {
		return legacySimDefaultPassword
	}
	return fallback
}
