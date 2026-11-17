package ucgctrl

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"hello/internal/platform/httpmeta"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// InternalMediaUploadApk POST /ucg/internal/api/media/upload-apk — gateway 版本管理 APK 直传 OSS。
//
// 业务：与 logo/视频上传隔离；上限约 200MB；multipart 大文件落临时盘后流式 PutObject，不整包进堆。
// 鉴权：DEVICE_GATEWAY_INTERNAL_SECRET（与其它 internal media 一致）。
// 表单字段：file 或 apk；扩展名必须 .apk。
func InternalMediaUploadApk(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	if !requireUcgInternalSecret(r) {
		return
	}
	maxBytes := ucgsvc.MaxApkUploadBytes(ctx)
	// maxMemory 32MB：更大文件由 Go 写临时文件，再 FormFile 打开流式读取。
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		glog.Warningf(ctx, "[ucg-internal-apk] ParseMultipartForm 失败 err=%v", err)
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "multipart 解析失败"))
		return
	}
	file, hdr, err := r.Request.FormFile("file")
	if err != nil || file == nil {
		file, hdr, err = r.Request.FormFile("apk")
	}
	if err != nil || file == nil {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 file 或 apk 字段"))
		return
	}
	defer file.Close()

	name := ""
	var size int64
	if hdr != nil {
		name = strings.TrimSpace(hdr.Filename)
		size = hdr.Size
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	if ext == "" {
		ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(r.GetForm("extension").String())), ".")
	}
	if ext != "apk" {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "仅支持 .apk 文件"))
		return
	}
	if size <= 0 {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "无法确定 APK 大小"))
		return
	}
	if size > maxBytes {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "APK 文件过大"))
		return
	}

	objectKey, cdnURL, err := ucgsvc.UploadApkObject(ctx, io.LimitReader(file, size), size)
	if err != nil {
		glog.Warningf(ctx, "[ucg-internal-apk] 上传失败 size=%d err=%v", size, err)
		writeUcgInternalUploadFail(r, err)
		return
	}
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"data": g.Map{
			"objectKey": objectKey,
			"cdnUrl":    cdnURL,
		},
	})
}

// InternalMediaDeleteApk POST /ucg/internal/api/media/delete-apk — 按 objectKey 删除 apk/ 对象。
//
// Body JSON: {"objectKey":"apk/..."}；无 ownership 校验（版本 APK 非用户媒资）。
func InternalMediaDeleteApk(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	if !requireUcgInternalSecret(r) {
		return
	}
	var body struct {
		ObjectKey string `json:"objectKey"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err != nil {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "JSON 解析失败"))
		return
	}
	key := strings.TrimSpace(body.ObjectKey)
	if key == "" {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "objectKey 不能为空"))
		return
	}
	if err := ucgsvc.DeleteApkObject(ctx, key); err != nil {
		glog.Warningf(ctx, "[ucg-internal-apk] 删除失败 key=%s err=%v", key, err)
		writeUcgInternalUploadFail(r, err)
		return
	}
	r.Response.WriteJson(g.Map{"code": 0, "message": "ok"})
}

// requireUcgInternalSecret 校验内部网关密钥；失败时已写 403。
func requireUcgInternalSecret(r *ghttp.Request) bool {
	secret := strings.TrimSpace(r.GetHeader(httpmeta.HeaderDeviceGatewayInternalSecret))
	if secret == "" {
		secret = strings.TrimSpace(r.GetHeader("X-Gateway-Internal-Secret"))
	}
	if !httpmeta.ValidateInternalSecret(secret) {
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "内部接口未授权"})
		r.ExitAll()
		return false
	}
	return true
}
