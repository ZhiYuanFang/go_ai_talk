package ucg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

const polishUserPrompt = "作为宝宝家长，你正在发朋友圈，选择了这些图，说点什么吧。"

const polishMaxTokens = 1024

// PolishPostText calls DashScope Qwen vision (OpenAI-compatible chat) to polish compose text from uploaded image keys.
func PolishPostText(ctx context.Context, imageKeys []string, draftText string) (string, error) {
	cfg := LoadAIConfig(ctx)
	if strings.TrimSpace(cfg.DashScopeAPIKey) == "" {
		return "", gerror.NewCode(gcode.CodeNotImplemented, "AI 润笔未配置")
	}
	if len(imageKeys) == 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "imageKeys 不能为空")
	}
	maxN := cfg.MaxImagesPerRequest
	if maxN <= 0 {
		maxN = 9
	}
	if len(imageKeys) > maxN {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, fmt.Sprintf("图片数量超过上限 %d", maxN))
	}

	contentParts := make([]map[string]interface{}, 0, len(imageKeys)+1)
	for _, key := range imageKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		cdnURL := BuildCdnURL(key)
		if cdnURL == "" {
			return "", gerror.NewCode(gcode.CodeInvalidParameter, "图片 CDN URL 无效")
		}
		contentParts = append(contentParts, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": cdnURL,
			},
		})
	}
	if len(contentParts) == 0 {
		return "", gerror.NewCode(gcode.CodeInvalidParameter, "无有效图片 URL")
	}

	prompt := polishUserPrompt
	if draft := strings.TrimSpace(draftText); draft != "" {
		prompt = "作为宝宝家长，你正在发朋友圈，选择了这些图。用户当前草稿：\n" + draft + "\n\n请结合图片润色或重写为适合发布的正文。"
	}
	contentParts = append(contentParts, map[string]interface{}{
		"type": "text",
		"text": prompt,
	})

	payload := map[string]interface{}{
		"model": cfg.VisionModel,
		"messages": []map[string]interface{}{
			{"role": "user", "content": contentParts},
		},
		"max_tokens": polishMaxTokens,
		"stream":     false,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "请求体序列化失败")
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodPost, cfg.VisionEndpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "创建请求失败")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.DashScopeAPIKey)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeOperationFailed, err, "AI 服务请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "读取 AI 响应失败")
	}
	if resp.StatusCode >= 300 {
		if resp.StatusCode == 429 {
			return "", gerror.NewCode(gcode.CodeOperationFailed, "AI 服务繁忙，请稍后再试")
		}
		msg := fmt.Sprintf("AI 服务错误: %d", resp.StatusCode)
		if detail := strings.TrimSpace(string(body)); detail != "" && len(detail) <= 512 {
			msg = msg + " — " + detail
		}
		return "", gerror.NewCode(gcode.CodeOperationFailed, msg)
	}

	text, err := extractPolishReply(body)
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", gerror.NewCode(gcode.CodeOperationFailed, "AI 未返回有效正文")
	}
	return text, nil
}

func extractPolishReply(body []byte) (string, error) {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", gerror.WrapCode(gcode.CodeInternalError, err, "解析 AI 响应失败")
	}
	if len(parsed.Choices) == 0 {
		return "", gerror.NewCode(gcode.CodeOperationFailed, "AI 响应为空")
	}
	return parsed.Choices[0].Message.Content, nil
}
