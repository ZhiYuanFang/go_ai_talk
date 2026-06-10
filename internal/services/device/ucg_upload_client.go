package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	ucgServiceBaseURLEnv       = "UCG_SERVICE_BASE_URL"
	defaultUcgServiceBaseURL   = "http://127.0.0.1:9804"
	ucgInternalMediaUploadPath = "/ucg/internal/api/media/upload"
)

// ucgUploadClient device-service 调用 ucg internal OSS 上传。
type ucgUploadClient struct {
	base   string
	secret string
	client *http.Client
}

var (
	ucgUploadOnce sync.Once
	ucgUploadInst *ucgUploadClient
)

func ucgUpload() *ucgUploadClient {
	ucgUploadOnce.Do(func() {
		ucgUploadInst = &ucgUploadClient{
			base:   resolveUcgServiceBaseURL(),
			secret: resolveDeviceGatewayInternalSecret(),
			client: &http.Client{Timeout: 60 * time.Second},
		}
	})
	return ucgUploadInst
}

func resolveUcgServiceBaseURL() string {
	if v := strings.TrimRight(strings.TrimSpace(os.Getenv(ucgServiceBaseURLEnv)), "/"); v != "" {
		return v
	}
	ctx := context.Background()
	if v := strings.TrimRight(strings.TrimSpace(g.Cfg().MustGet(ctx, "device.ucgServiceUrl").String()), "/"); v != "" {
		return v
	}
	return defaultUcgServiceBaseURL
}

func resolveDeviceGatewayInternalSecret() string {
	for _, key := range []string{"DEVICE_GATEWAY_INTERNAL_SECRET", "GATEWAY_INTERNAL_SECRET"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return strings.TrimSpace(g.Cfg().MustGet(context.Background(), "gatewayApp.deviceInternalSecret").String())
}

// uploadEventLogoViaUcg multipart 上传 logo 至 ucg internal，返回 objectKey。
func (c *ucgUploadClient) uploadEventLogoViaUcg(ctx context.Context, filename string, body []byte) (objectKey string, err error) {
	if c.base == "" {
		return "", fmt.Errorf("UCG_SERVICE_BASE_URL 未配置")
	}
	if c.secret == "" {
		return "", fmt.Errorf("DEVICE_GATEWAY_INTERNAL_SECRET 未配置")
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("logo", filename)
	if err != nil {
		return "", err
	}
	if _, err = part.Write(body); err != nil {
		return "", err
	}
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+ucgInternalMediaUploadPath, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set(HeaderDeviceGatewayInternalSecret, c.secret)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ucg 上传请求失败: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ObjectKey string `json:"objectKey"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return "", fmt.Errorf("ucg 响应解析失败: %w", err)
	}
	if envelope.Code != 0 {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = string(raw)
		}
		return "", fmt.Errorf("ucg 上传失败: %s", msg)
	}
	objectKey = strings.TrimSpace(envelope.Data.ObjectKey)
	if objectKey == "" {
		return "", fmt.Errorf("ucg 未返回 objectKey")
	}
	return objectKey, nil
}
