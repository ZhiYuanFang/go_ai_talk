package ucg

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"hello/internal/platform/httpmeta"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	ucgInternalApkUploadPath = "/ucg/internal/api/media/upload-apk"
	ucgInternalApkDeletePath = "/ucg/internal/api/media/delete-apk"
	// apkUploadTimeout 内网大文件直传预留（约 200MB）；显著长于 logo 60s。
	apkUploadTimeout = 15 * time.Minute
	apkDeleteTimeout = 60 * time.Second
)

// UploadApk 经 ucg internal 流式上传版本 APK，返回 objectKey 与 CDN URL。
//
// 业务：gateway-app 版本管理；MUST 使用 DEVICE_GATEWAY_INTERNAL_SECRET；
// UCG_SERVICE_URL 优先，缺省回退本机 ucg 端口。body 按 size 限读后写入 multipart，避免无界缓冲。
func UploadApk(ctx context.Context, filename string, body io.Reader, size int64) (objectKey, cdnURL string, err error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_BASE_URL")), "/")
	}
	if base == "" {
		base = defaultUcgServiceURL
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	if body == nil || size <= 0 {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "APK 内容无效")
	}
	name := strings.TrimSpace(filename)
	if name == "" {
		name = "app.apk"
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	contentType := mw.FormDataContentType()
	go func() {
		var copyErr error
		defer func() {
			_ = mw.Close()
			if copyErr != nil {
				_ = pw.CloseWithError(copyErr)
			} else {
				_ = pw.Close()
			}
		}()
		part, pErr := mw.CreateFormFile("file", name)
		if pErr != nil {
			copyErr = pErr
			return
		}
		_, copyErr = io.Copy(part, io.LimitReader(body, size))
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+ucgInternalApkUploadPath, pr)
	if err != nil {
		_ = pw.CloseWithError(err)
		return "", "", err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)

	client := &http.Client{Timeout: apkUploadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "ucg APK 上传失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var envelope struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ObjectKey string `json:"objectKey"`
			CdnURL    string `json:"cdnUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "ucg APK 响应解析失败")
	}
	if envelope.Code != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = string(raw)
		}
		return "", "", gerror.NewCode(gcode.CodeInternalError, "ucg APK 上传失败: "+msg)
	}
	objectKey = strings.TrimSpace(envelope.Data.ObjectKey)
	cdnURL = strings.TrimSpace(envelope.Data.CdnURL)
	if objectKey == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "ucg 未返回 objectKey")
	}
	return objectKey, cdnURL, nil
}

// DeleteApkObject 经 ucg internal 删除 apk/ 前缀对象（尽力调用；失败由调用方记 WARN）。
func DeleteApkObject(ctx context.Context, objectKey string) error {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		base = strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_BASE_URL")), "/")
	}
	if base == "" {
		base = defaultUcgServiceURL
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "objectKey 不能为空")
	}
	payload, _ := json.Marshal(map[string]string{"objectKey": key})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+ucgInternalApkDeletePath, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)

	client := &http.Client{Timeout: apkDeleteTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "ucg APK 删除失败")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var envelope struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "ucg APK 删除响应解析失败")
	}
	if envelope.Code != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = string(raw)
		}
		return gerror.NewCode(gcode.CodeInternalError, "ucg APK 删除失败: "+msg)
	}
	return nil
}
