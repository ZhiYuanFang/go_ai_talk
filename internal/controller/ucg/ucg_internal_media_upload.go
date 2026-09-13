package ucgctrl

import (
	"hello/internal/platform/httpmeta"
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ucgInternalMediaUpload POST /ucg/internal/api/media/upload — device 管理端事件 logo 等服务端直传 OSS。
func InternalMediaUpload(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	secret := strings.TrimSpace(r.GetHeader(httpmeta.HeaderDeviceGatewayInternalSecret))
	if secret == "" {
		secret = strings.TrimSpace(r.GetHeader("X-Gateway-Internal-Secret"))
	}
	if !httpmeta.ValidateInternalSecret(secret) {
		r.Response.Status = 403
		r.Response.WriteJson(g.Map{"code": 403, "message": "内部接口未授权"})
		r.ExitAll()
		return
	}
	if err := r.ParseMultipartForm(ucgsvc.MaxEventLogoBytes + (1 << 20)); err != nil {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "multipart 解析失败"))
		return
	}
	file, hdr, err := r.Request.FormFile("file")
	if err != nil || file == nil {
		// 兼容管理端字段名 logo
		file, hdr, err = r.Request.FormFile("logo")
	}
	if err != nil || file == nil {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 file 或 logo 字段"))
		return
	}
	defer file.Close()

	ext := strings.TrimSpace(r.GetForm("extension").String())
	if ext == "" && hdr != nil {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(hdr.Filename)), ".")
	}
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")

	// kind：event（默认）| feature；决定 OSS objectKey 前缀，兼容旧 device 上传不传 kind。
	kind := strings.ToLower(strings.TrimSpace(r.GetForm("kind").String()))
	if kind == "" {
		kind = "event"
	}

	limited := io.LimitReader(file, ucgsvc.MaxEventLogoBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		writeUcgInternalUploadFail(r, gerror.WrapCode(gcode.CodeInvalidParameter, err, "读取上传文件失败"))
		return
	}
	if len(data) > ucgsvc.MaxEventLogoBytes {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "logo 文件过大"))
		return
	}
	var objectKey, cdnURL string
	switch kind {
	case "feature":
		objectKey, cdnURL, err = ucgsvc.UploadFeatureLogoObject(ctx, ext, bytes.NewReader(data), int64(len(data)))
	case "event":
		objectKey, cdnURL, err = ucgsvc.UploadEventLogoObject(ctx, ext, bytes.NewReader(data), int64(len(data)))
	default:
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "kind 须为 event 或 feature"))
		return
	}
	if err != nil {
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

func writeUcgInternalUploadFail(r *ghttp.Request, err error) {
	if err == nil {
		return
	}
	bizCode := gcode.CodeInternalError.Code()
	msg := err.Error()
	if ge, ok := err.(*gerror.Error); ok && ge.Code() != gcode.CodeNil {
		bizCode = ge.Code().Code()
		if m := ge.Error(); m != "" {
			msg = m
		}
	}
	r.Response.WriteJson(g.Map{"code": bizCode, "message": msg})
}
