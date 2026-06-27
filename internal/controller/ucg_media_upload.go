package controller

import (
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

// ucgMediaUpload POST /ucg/app/api/media/upload — multipart 同域代理上传（Web 规避 OSS CORS）。
func ucgMediaUpload(r *ghttp.Request) {
	if r.Method != http.MethodPost {
		r.Response.WriteStatusExit(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	wxID, err := wxIDFromUcgHeader(r)
	if err != nil {
		writeUcgMediaUploadFail(r, err)
		return
	}
	if err := r.ParseMultipartForm(ucgsvc.MaxMediaUploadBytes + (1 << 20)); err != nil {
		writeUcgMediaUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "multipart 解析失败"))
		return
	}
	file, hdr, err := r.Request.FormFile("file")
	if err != nil || file == nil {
		writeUcgMediaUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 file 字段"))
		return
	}
	defer file.Close()

	mediaKind := r.GetForm("mediaKind").Int()
	if mediaKind != 1 && mediaKind != 2 {
		writeUcgMediaUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "mediaKind 须为 1 或 2"))
		return
	}

	ext := strings.TrimSpace(r.GetForm("extension").String())
	if ext == "" && hdr != nil {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(hdr.Filename)), ".")
	}
	ext = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")

	limited := io.LimitReader(file, ucgsvc.MaxMediaUploadBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		writeUcgMediaUploadFail(r, gerror.WrapCode(gcode.CodeInvalidParameter, err, "读取上传文件失败"))
		return
	}
	if len(data) > ucgsvc.MaxMediaUploadBytes {
		writeUcgMediaUploadFail(r, gerror.NewCode(gcode.CodeInvalidParameter, "文件超过大小上限"))
		return
	}
	result, err := ucgsvc.UploadMediaObject(ctx, wxID, mediaKind, ext, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		writeUcgMediaUploadFail(r, err)
		return
	}
	dataMap := g.Map{
		"objectKey": result.ObjectKey,
		"cdnUrl":    result.CdnURL,
	}
	if mediaKind == 2 {
		if strings.TrimSpace(result.ContentHash) != "" {
			dataMap["contentHash"] = result.ContentHash
		}
		if strings.TrimSpace(result.TransformVersion) != "" {
			dataMap["transformVersion"] = result.TransformVersion
		}
	}
	r.Response.WriteJson(g.Map{
		"code":    0,
		"message": "ok",
		"data":    dataMap,
	})
}

func writeUcgMediaUploadFail(r *ghttp.Request, err error) {
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
