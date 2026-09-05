package ucgctrl

import (
	"hello/internal/platform/httpmeta"
	"io"
	"net/http"
	"strings"

	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ucgInternalMediaUploadVideo POST /ucg/internal/api/media/upload-video — 服务端转码后上传 OSS（sim T4 等）。
func InternalMediaUploadVideo(r *ghttp.Request) {
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
	if err := r.ParseMultipartForm(ucgsvc.MaxMediaUploadBytes + (1 << 20)); err != nil {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "multipart 解析失败"))
		return
	}
	file, hdr, err := r.Request.FormFile("file")
	if err != nil || file == nil {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 file 字段"))
		return
	}
	defer file.Close()

	filename := "video.mp4"
	if hdr != nil && strings.TrimSpace(hdr.Filename) != "" {
		filename = hdr.Filename
	}

	limited := io.LimitReader(file, ucgsvc.MaxMediaUploadBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		writeUcgInternalUploadFail(r, gerror.WrapCode(gcode.CodeInvalidParameter, err, "读取上传文件失败"))
		return
	}
	if len(data) > ucgsvc.MaxMediaUploadBytes {
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "文件过大"))
		return
	}
	_ = filename

	result, err := ucgsvc.UploadVideoTranscodedObject(ctx, data)
	if err != nil {
		writeUcgInternalUploadFail(r, err)
		return
	}
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"data": g.Map{
			"objectKey":   result.ObjectKey,
			"cdnUrl":      result.CdnURL,
			"contentHash": result.ContentHash,
		},
	})
}
