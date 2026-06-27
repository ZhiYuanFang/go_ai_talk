package simuser

import "strings"

// commentVisionUserContent 构建 T2 评论用 user message Content：有封面 CDN 时为多模态，否则纯文本。
func commentVisionUserContent(coverCdnURL, promptText string) interface{} {
	coverCdnURL = strings.TrimSpace(coverCdnURL)
	promptText = strings.TrimSpace(promptText)
	if coverCdnURL == "" {
		return promptText
	}
	return []map[string]interface{}{
		{
			"type": "image_url",
			"image_url": map[string]string{
				"url": coverCdnURL,
			},
		},
		{
			"type": "text",
			"text": promptText,
		},
	}
}
