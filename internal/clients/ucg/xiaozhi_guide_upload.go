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
	ucgInternalMediaUploadVideoPath = "/ucg/internal/api/media/upload-video"
	xiaozhiGuideUploadKind          = "xiaozhi-guide"
)

// UploadXiaozhiGuideVideo 经 ucg internal 上传小智引导视频（v1 验真直传，前缀 xiaozhi-guide/）。
//
// 业务：cash Hub Admin 覆盖双槽视频；与 Flutter/UCG 同体积与格式规则，无服务端转码。
// Args: filename 仅作 multipart 文件名；body 为完整视频字节。
// Returns: objectKey 与 CDN URL。
func UploadXiaozhiGuideVideo(ctx context.Context, filename string, body []byte) (objectKey, cdnURL string, err error) {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("UCG_SERVICE_URL")), "/")
	if base == "" {
		base = defaultUcgServiceURL
	}
	secret := strings.TrimSpace(os.Getenv("DEVICE_GATEWAY_INTERNAL_SECRET"))
	if secret == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "未配置 DEVICE_GATEWAY_INTERNAL_SECRET")
	}
	if len(body) == 0 {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "视频文件为空")
	}
	name := strings.TrimSpace(filename)
	if name == "" {
		name = "guide.mp4"
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", name)
	if err != nil {
		return "", "", err
	}
	if _, err = part.Write(body); err != nil {
		return "", "", err
	}
	if err = w.WriteField("kind", xiaozhiGuideUploadKind); err != nil {
		return "", "", err
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+ucgInternalMediaUploadVideoPath, &buf)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set(httpmeta.HeaderDeviceGatewayInternalSecret, secret)

	// 与 Flutter/UCG 同档视频；验真+PUT 可能较慢。
	client := &http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "ucg 引导视频上传失败")
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
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "ucg 上传失败: "+msg)
	}
	objectKey = strings.TrimSpace(envelope.Data.ObjectKey)
	cdnURL = strings.TrimSpace(envelope.Data.CdnURL)
	if objectKey == "" {
		return "", "", gerror.NewCode(gcode.CodeInternalError, "ucg 未返回 objectKey")
	}
	return objectKey, cdnURL, nil
}
