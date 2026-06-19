package simuser

import (
	"context"
	"strings"
)

// RenderPrompt 将模板变量替换为实际值。
func RenderPrompt(tmpl string, vars map[string]string) string {
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

// LoadRenderedPrompt 加载并渲染指定 taskType 的用户 prompt。
func LoadRenderedPrompt(ctx context.Context, taskType string, vars map[string]string) (system string, user string, err error) {
	p, err := GetPrompt(ctx, taskType)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(p.SystemPrompt), RenderPrompt(p.UserPromptTemplate, vars), nil
}
