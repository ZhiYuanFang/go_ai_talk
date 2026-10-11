package ucgctrl

import (
	"io"
	"net/http"
	"strings"

	"hello/internal/platform/httpmeta"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// InternalMediaUploadVideo POST /ucg/internal/api/media/upload-video — v1 验真后直传 OSS（无转码）。
//
// 业务：sim T4 / 运营引导片等内部上传；可选 form kind=xiaozhi-guide 写入专用前缀。
// Args: multipart file；可选 kind（空=默认社区前缀；xiaozhi-guide=引导片前缀）。
// Returns: objectKey、cdnUrl、contentHash、transformVersion=v1。
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

	_ = hdr
	kind := strings.ToLower(strings.TrimSpace(r.GetForm("kind").String()))

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

	var result *ucgsvc.UploadMediaResult
	switch kind {
	case "", "default":
		result, err = ucgsvc.UploadValidatedVideoObject(ctx, "", data)
	case "xiaozhi-guide":
		result, err = ucgsvc.UploadXiaozhiGuideVideoObject(ctx, data)
	default:
		writeUcgInternalUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "kind 须为空或 xiaozhi-guide"))
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
			"objectKey":         result.ObjectKey,
			"cdnUrl":            result.CdnURL,
			"contentHash":       result.ContentHash,
			"transformVersion":  result.TransformVersion,
		},
	})
}
