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
	ucgInternalMediaUploadPath = "/ucg/internal/api/media/upload"
	defaultUcgServiceURL       = "http://127.0.0.1:9804"
)

// UploadFeatureLogo 经 ucg internal 上传功能 logo（kind=feature），返回 objectKey 与 CDN URL。
//
// 业务：cash Admin 上传功能品牌图；MUST 使用 DEVICE_GATEWAY_INTERNAL_SECRET；
// UCG_SERVICE_URL 优先，缺省回退本机 ucg 端口。
func UploadFeatureLogo(ctx context.Context, filename string, body []byte) (objectKey, cdnURL string, err error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		base = defaultUcgServiceURL
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	if len(body) == 0 {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "logo 文件为空")
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("logo", filename)
	if err != nil {
		return "", "", err
	}
	if _, err = part.Write(body); err != nil {
		return "", "", err
	}
	// kind=feature → ucg 写入 feature/ 前缀
	if err = w.WriteField("kind", "feature"); err != nil {
		return "", "", err
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+ucgInternalMediaUploadPath, &buf)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "ucg 功能 logo 上传失败")
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
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "ucg 响应解析失败")
	}
	if envelope.Code != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = string(raw)
		}
		return "", "", gerror.NewCode(gcode.CodeInternalError, "ucg 上传失败: "+msg)
	}
	objectKey = strings.TrimSpace(envelope.Data.ObjectKey)
	cdnURL = strings.TrimSpace(envelope.Data.CdnURL)
	if objectKey == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "ucg 未返回 objectKey")
	}
	return objectKey, cdnURL, nil
}
