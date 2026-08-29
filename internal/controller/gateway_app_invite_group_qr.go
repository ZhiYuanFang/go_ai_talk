package controller

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

const inviteGroupQrFileName = "er_code.png"

// gatewayAppInviteGroupQrUpload POST：Admin 上传微信群二维码，强制覆盖为 er_code.png。
func gatewayAppInviteGroupQrUpload(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	if !requireGatewayAdminHTTP(r) {
		return
	}
	// 图片上限 5MB（群二维码无需 APK 级体积）。
	const maxBytes int64 = 5 << 20
	if err := r.ParseMultipartForm(maxBytes + (1 << 20)); err != nil {
		glog.Warningf(ctx, "[invite-group-qr] ParseMultipartForm 失败 err=%v", err)
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "multipart 解析失败"})
		return
	}
	file, hdr, err := r.Request.FormFile("file")
	if err != nil {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "请选择 PNG 文件（字段名 file）"})
		return
	}
	defer func() { _ = file.Close() }()
	if hdr.Size > 0 && hdr.Size > maxBytes {
		r.Response.Status = http.StatusRequestEntityTooLarge
		r.Response.WriteJson(g.Map{"code": 413, "message": "文件过大，上限 5MB"})
		return
	}
	orig := strings.ToLower(strings.TrimSpace(hdr.Filename))
	if !strings.HasSuffix(orig, ".png") {
		r.Response.Status = http.StatusBadRequest
		r.Response.WriteJson(g.Map{"code": 400, "message": "仅支持 .png"})
		return
	}
	dir := filepath.Clean(gatewayapp.ApkStorageDir(ctx))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		glog.Warningf(ctx, "[invite-group-qr] mkdir 失败 dir=%s err=%v", dir, err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "创建存储目录失败"})
		return
	}
	dest := filepath.Join(dir, inviteGroupQrFileName)
	out, err := os.Create(dest)
	if err != nil {
		glog.Warningf(ctx, "[invite-group-qr] create 失败 path=%s err=%v", dest, err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "保存文件失败"})
		return
	}
	defer func() { _ = out.Close() }()
	written, err := io.Copy(out, io.LimitReader(file, maxBytes+1))
	if err != nil {
		_ = os.Remove(dest)
		glog.Warningf(ctx, "[invite-group-qr] write 失败 err=%v", err)
		r.Response.Status = http.StatusInternalServerError
		r.Response.WriteJson(g.Map{"code": 500, "message": "写入文件失败"})
		return
	}
	if written > maxBytes {
		_ = os.Remove(dest)
		r.Response.Status = http.StatusRequestEntityTooLarge
		r.Response.WriteJson(g.Map{"code": 413, "message": "文件过大，上限 5MB"})
		return
	}
	// 校验 PNG 魔数，避免改扩展名绕过。
	if written >= 8 {
		if f, e := os.Open(dest); e == nil {
			buf := make([]byte, 8)
			_, _ = f.Read(buf)
			_ = f.Close()
			pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
			ok := true
			for i := 0; i < 8; i++ {
				if buf[i] != pngMagic[i] {
					ok = false
					break
				}
			}
			if !ok {
				_ = os.Remove(dest)
				r.Response.Status = http.StatusBadRequest
				r.Response.WriteJson(g.Map{"code": 400, "message": "文件不是有效 PNG"})
				return
			}
		}
	}
	dlPath := versionAdminApkURLPrefix + inviteGroupQrFileName
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"data": g.Map{
			"path":     dlPath,
			"fileName": inviteGroupQrFileName,
		},
	})
}
