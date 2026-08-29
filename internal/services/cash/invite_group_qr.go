package cash

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	// InviteGroupQrFileName 网关 APK 目录内固定文件名。
	InviteGroupQrFileName = "er_code.png"
	// InviteGroupQrPath 对外相对路径（经 gateway 匿名下载）。
	InviteGroupQrPath = "/device/app/apk/" + InviteGroupQrFileName
	inviteGroupQrRowID = 1
)

// InviteGroupQrAdmin 管理端读模型（含过期仍可预览）。
type InviteGroupQrAdmin struct {
	FileName   string `json:"fileName"`
	ExpiresAt  int64  `json:"expiresAt"`
	UpdatedAt  int64  `json:"updatedAt"`
	PreviewPath string `json:"previewPath"` // 相对 path + ?v=
	AppVisible bool   `json:"appVisible"`  // 当前是否会对 App catalog 返回
}

// GetInviteGroupQrAdmin 读取 singleton；无行则返回零值模板。
func GetInviteGroupQrAdmin(ctx context.Context) (*InviteGroupQrAdmin, error) {
	var row struct {
		ExpiresAt int64  `json:"expires_at"`
		UpdatedAt int64  `json:"updated_at"`
		FileName  string `json:"file_name"`
	}
	_ = g.DB().Model("invite_group_qr").Ctx(ctx).Where("id", inviteGroupQrRowID).Scan(&row)
	fn := strings.TrimSpace(row.FileName)
	if fn == "" {
		fn = InviteGroupQrFileName
	}
	now := time.Now().Unix()
	out := &InviteGroupQrAdmin{
		FileName:    fn,
		ExpiresAt:   row.ExpiresAt,
		UpdatedAt:   row.UpdatedAt,
		PreviewPath: buildInviteGroupQrRelURL(row.UpdatedAt),
		AppVisible:  row.ExpiresAt > 0 && row.ExpiresAt > now,
	}
	return out, nil
}

// UpsertInviteGroupQrMeta 更新有效期；touchUpdated 为真时刷新 updated_at（上传后调用）。
func UpsertInviteGroupQrMeta(ctx context.Context, expiresAt int64, touchUpdated bool) error {
	if expiresAt < 0 {
		expiresAt = 0
	}
	now := time.Now().Unix()
	n, err := g.DB().Model("invite_group_qr").Ctx(ctx).Where("id", inviteGroupQrRowID).Count()
	if err != nil {
		return err
	}
	data := g.Map{
		"file_name":  InviteGroupQrFileName,
		"expires_at": expiresAt,
	}
	if touchUpdated || n == 0 {
		data["updated_at"] = now
	}
	if n == 0 {
		data["id"] = inviteGroupQrRowID
		if _, ok := data["updated_at"]; !ok {
			data["updated_at"] = now
		}
		_, err = g.DB().Model("invite_group_qr").Ctx(ctx).Data(data).Insert()
		return err
	}
	_, err = g.DB().Model("invite_group_qr").Ctx(ctx).Where("id", inviteGroupQrRowID).Data(data).Update()
	return err
}

// TouchInviteGroupQrUpdated 上传成功后刷新 updated_at（保留原 expires_at）。
func TouchInviteGroupQrUpdated(ctx context.Context) error {
	now := time.Now().Unix()
	n, err := g.DB().Model("invite_group_qr").Ctx(ctx).Where("id", inviteGroupQrRowID).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		_, err = g.DB().Model("invite_group_qr").Ctx(ctx).Data(g.Map{
			"id": inviteGroupQrRowID, "file_name": InviteGroupQrFileName,
			"expires_at": 0, "updated_at": now,
		}).Insert()
		return err
	}
	_, err = g.DB().Model("invite_group_qr").Ctx(ctx).Where("id", inviteGroupQrRowID).Data(g.Map{
		"file_name": InviteGroupQrFileName, "updated_at": now,
	}).Update()
	return err
}

// ResolveInviteGroupQrURLForApp 仅当未过期时返回可给 App 的 URL（绝对优先，否则相对）。
func ResolveInviteGroupQrURLForApp(ctx context.Context) (string, error) {
	var row struct {
		ExpiresAt int64 `json:"expires_at"`
		UpdatedAt int64 `json:"updated_at"`
	}
	err := g.DB().Model("invite_group_qr").Ctx(ctx).Where("id", inviteGroupQrRowID).Scan(&row)
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	if row.ExpiresAt <= 0 || row.ExpiresAt <= now {
		return "", nil
	}
	rel := buildInviteGroupQrRelURL(row.UpdatedAt)
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GATEWAY_APP_PUBLIC_BASE_URL")), "/")
	if base == "" {
		return rel, nil
	}
	return base + rel, nil
}

func buildInviteGroupQrRelURL(updatedAt int64) string {
	if updatedAt <= 0 {
		return InviteGroupQrPath
	}
	return fmt.Sprintf("%s?v=%d", InviteGroupQrPath, updatedAt)
}

// SetInviteGroupQrExpires 仅改有效期（Admin 保存）。
func SetInviteGroupQrExpires(ctx context.Context, expiresAt int64) error {
	if expiresAt < 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "expiresAt 无效")
	}
	return UpsertInviteGroupQrMeta(ctx, expiresAt, false)
}
