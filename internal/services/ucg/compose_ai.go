package ucg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

const polishUserPrompt = "作为宝宝家长，你正在发朋友圈，选择了这些图，说点什么吧。"

const polishMaxTokens = 1024

// PolishPostText 经 LanePolish 调用多模态 LLM 润笔正文。
func PolishPostText(ctx context.Context, imageKeys []string, draftText string) (string, error) {
	cfg := LoadAIConfig(ctx)
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

	resp, err := aimodel.Invoke(ctx, aimodel.LanePolish, aimodel.ChatRequest{
		Messages: []aimodel.Message{
			{Role: "user", Content: contentParts},
		},
		MaxTokens:  polishMaxTokens,
		TimeoutSec: cfg.TimeoutSeconds,
	})
	if err != nil {
		return "", mapPolishLLMError(err)
	}
	text := strings.TrimSpace(resp.Content)
	if text == "" {
		if t, pErr := extractPolishReply(resp.RawBody); pErr == nil {
			text = strings.TrimSpace(t)
		}
	}
	if text == "" {
		return "", gerror.NewCode(gcode.CodeOperationFailed, "AI 未返回有效正文")
	}
	return text, nil
}

func mapPolishLLMError(err error) error {
	if aimodel.IsQueueFull(err) {
		return gerror.NewCode(gcode.New(aimodel.CodeLLMQueueFull, err.Error(), nil), err.Error())
	}
	if errors.Is(err, aimodel.ErrProviderKeyMissing) {
		return gerror.NewCode(gcode.CodeNotImplemented, "AI 润笔未配置")
	}
	return gerror.WrapCode(gcode.CodeOperationFailed, err, "AI 服务请求失败")
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
