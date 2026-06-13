// Package ucg 阿里云 Green 内容安全客户端。
//
// 【与 Green 风暴相关的返回值语义】
// - ModerateText/Image 返回 err：上层视为 Phase1 失败 → MQ requeue → 可能重复调 API
// - 返回 Pass=false：违规，上层 persist reject，handler 成功 Ack
// - parseTextModeration 中 body.Code!=200（如额度 588）→ err，控制台可能有请求记录但 verdict 不落库
package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green "github.com/alibabacloud-go/green-20220302/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	rejectReasonDefault = "违规已下架"
	greenDataIDMaxLen   = 64 // 阿里云 ImageModeration/VideoModeration dataId 上限
)

// greenDataIDFromMediaURL 从 CDN URL 的 path 推导合规 dataId（≤64，字符集 A-Za-z0-9_.-）。
// 完整 HTTP(S) URL 含 : / 等非法字符且常超 64 字符，不得作为 dataId；path 与 objectKey 一致（如 social/2026/06/xxx.jpg）。
// 无法推导时返回空串，调用方 MUST 省略 dataId 字段而非传非法值。
func greenDataIDFromMediaURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range path {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-', r == '.':
			b.WriteRune(r)
		case r == '/':
			b.WriteRune('_')
		}
	}
	id := b.String()
	if id == "" {
		return ""
	}
	if len(id) > greenDataIDMaxLen {
		id = id[:greenDataIDMaxLen]
	}
	return id
}

// greenServiceParamsJSON 序列化 ServiceParameters；mediaURL 非空时尝试附加合规 dataId。
func greenServiceParamsJSON(base map[string]interface{}, mediaURL string) (string, error) {
	if dataID := greenDataIDFromMediaURL(mediaURL); dataID != "" {
		base["dataId"] = dataID
	}
	b, err := json.Marshal(base)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AuditVerdict Green 单次机审结论；Pass=false 为内容违规（非 API 错误）。
type AuditVerdict struct {
	Pass   bool
	Reason string
}

// GreenModerator 内容审核接口（帖子/资料/评论/私信共用）。
type GreenModerator interface {
	ModerateText(ctx context.Context, service, content string) (AuditVerdict, error)
	ModerateImageURL(ctx context.Context, imageURL string) (AuditVerdict, error)
	ModerateVideoURL(ctx context.Context, videoURL string) (AuditVerdict, error)
	Enabled() bool // true=真实调阿里云；false=noop
}

type greenModerator struct {
	client  *green.Client
	enabled bool
}

var (
	greenOnce sync.Once // 进程内单例；改 ucg.green.enabled 需 restart
	greenIns  GreenModerator
)

// Green 返回 Green 客户端单例；enabled=false 或建连失败时用 noop（不调 API、全 pass）。
func Green() GreenModerator {
	greenOnce.Do(func() {
		cfg := LoadGreenConfig(context.Background())
		if !cfg.Enabled {
			greenIns = &noopGreenModerator{}
			return
		}
		client, err := newGreenClient(cfg)
		if err != nil {
			greenIns = &noopGreenModerator{}
			return
		}
		greenIns = &greenModerator{client: client, enabled: true}
	})
	return greenIns
}

type noopGreenModerator struct{}

func (n *noopGreenModerator) Enabled() bool { return false }

func (n *noopGreenModerator) ModerateText(ctx context.Context, service, content string) (AuditVerdict, error) {
	_ = ctx
	_ = service
	_ = content
	return AuditVerdict{Pass: true}, nil // 不调 API，直接 pass
}

func (n *noopGreenModerator) ModerateImageURL(ctx context.Context, imageURL string) (AuditVerdict, error) {
	_ = ctx
	_ = imageURL
	return AuditVerdict{Pass: true}, nil
}

func (n *noopGreenModerator) ModerateVideoURL(ctx context.Context, videoURL string) (AuditVerdict, error) {
	_ = ctx
	_ = videoURL
	return AuditVerdict{Pass: true}, nil
}

func (g *greenModerator) Enabled() bool { return g.enabled }

func newGreenClient(cfg GreenConfig) (*green.Client, error) {
	conf := &openapi.Config{
		AccessKeyId:     tea.String(cfg.AccessKeyID),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
		Endpoint:        tea.String(cfg.Endpoint),
		ConnectTimeout:  tea.Int(3000),
		ReadTimeout:     tea.Int(10000),
	}
	return green.NewClient(conf)
}

// ModerateText 文本机审；service 常用 nickname_detection / comment_detection。
func (g *greenModerator) ModerateText(ctx context.Context, service, content string) (AuditVerdict, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return AuditVerdict{Pass: true}, nil
	}
	if service == "" {
		service = "comment_detection"
	}
	params, err := json.Marshal(map[string]interface{}{"content": content})
	if err != nil {
		return AuditVerdict{}, err
	}
	req := &green.TextModerationRequest{
		Service:           tea.String(service),
		ServiceParameters: tea.String(string(params)),
	}
	resp, err := g.client.TextModeration(req)
	if err != nil {
		// SDK/网络错误 → err → Phase1 失败 → requeue → 风暴
		return AuditVerdict{}, err
	}
	return parseTextModeration(resp)
}

func (g *greenModerator) ModerateImageURL(ctx context.Context, imageURL string) (AuditVerdict, error) {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return AuditVerdict{Pass: true}, nil
	}
	params, err := greenServiceParamsJSON(map[string]interface{}{
		"imageUrl": imageURL,
	}, imageURL)
	if err != nil {
		return AuditVerdict{}, err
	}
	req := &green.ImageModerationRequest{
		Service:           tea.String("baselineCheck"),
		ServiceParameters: tea.String(string(params)),
	}
	resp, err := g.client.ImageModeration(req)
	if err != nil {
		return AuditVerdict{}, err
	}
	return parseImageModeration(resp)
}

func (g *greenModerator) ModerateVideoURL(ctx context.Context, videoURL string) (AuditVerdict, error) {
	videoURL = strings.TrimSpace(videoURL)
	if videoURL == "" {
		return AuditVerdict{Pass: true}, nil
	}
	params, err := greenServiceParamsJSON(map[string]interface{}{
		"url": videoURL,
	}, videoURL)
	if err != nil {
		return AuditVerdict{}, err
	}
	req := &green.VideoModerationRequest{
		Service:           tea.String("videoDetection"),
		ServiceParameters: tea.String(string(params)),
	}
	resp, err := g.client.VideoModeration(req)
	if err != nil {
		return AuditVerdict{}, err
	}
	return parseVideoModeration(resp)
}

// normalizeGreenRejectReason 将 Green Data.Reason JSON 中的 riskTips 提取为用户可读文案。
func normalizeGreenRejectReason(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "{") {
		return raw
	}
	var parsed struct {
		RiskTips string `json:"riskTips"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.RiskTips)
}

// parseTextModeration 解析文本审核 HTTP 响应。
// body.Code!=200 时返回 error（非 Pass=false）→ 不落 verdict → 可能 Green 风暴。
func parseTextModeration(resp *green.TextModerationResponse) (AuditVerdict, error) {
	if resp == nil || resp.Body == nil {
		return AuditVerdict{}, fmt.Errorf("green text: empty response")
	}
	if resp.StatusCode == nil || *resp.StatusCode != http.StatusOK {
		code := int32(0)
		if resp.StatusCode != nil {
			code = *resp.StatusCode
		}
		return AuditVerdict{}, fmt.Errorf("green text: http %d", code)
	}
	if resp.Body.Code == nil || *resp.Body.Code != 200 {
		code := int32(0)
		if resp.Body.Code != nil {
			code = *resp.Body.Code
		}
		// 额度不足/限流等常见为非 200 business code → err → requeue
		return AuditVerdict{}, fmt.Errorf("green text: code %d", code)
	}
	labels := strings.TrimSpace(tea.StringValue(resp.Body.Data.Labels))
	if labels == "" || labels == "nonLabel" {
		return AuditVerdict{Pass: true}, nil
	}
	reason := normalizeGreenRejectReason(tea.StringValue(resp.Body.Data.Reason))
	if reason == "" {
		reason = rejectReasonDefault
	}
	return AuditVerdict{Pass: false, Reason: reason}, nil
}

func parseImageModeration(resp *green.ImageModerationResponse) (AuditVerdict, error) {
	if resp == nil || resp.Body == nil {
		return AuditVerdict{}, fmt.Errorf("green image: empty response")
	}
	if resp.StatusCode == nil || *resp.StatusCode != http.StatusOK {
		code := int32(0)
		if resp.StatusCode != nil {
			code = *resp.StatusCode
		}
		return AuditVerdict{}, fmt.Errorf("green image: http %d", code)
	}
	if resp.Body.Code == nil || *resp.Body.Code != 200 {
		code := int32(0)
		if resp.Body.Code != nil {
			code = *resp.Body.Code
		}
		msg := strings.TrimSpace(tea.StringValue(resp.Body.Msg))
		if msg != "" {
			return AuditVerdict{}, fmt.Errorf("green image: code %d msg %s", code, msg)
		}
		return AuditVerdict{}, fmt.Errorf("green image: code %d", code)
	}
	for _, r := range resp.Body.Data.Result {
		label := strings.TrimSpace(tea.StringValue(r.Label))
		if label != "" && label != "nonLabel" && label != "normal" {
			reason := rejectReasonDefault
			return AuditVerdict{Pass: false, Reason: reason}, nil
		}
	}
	return AuditVerdict{Pass: true}, nil
}

func parseVideoModeration(resp *green.VideoModerationResponse) (AuditVerdict, error) {
	if resp == nil || resp.Body == nil {
		return AuditVerdict{}, fmt.Errorf("green video: empty response")
	}
	if resp.StatusCode == nil || *resp.StatusCode != http.StatusOK {
		code := int32(0)
		if resp.StatusCode != nil {
			code = *resp.StatusCode
		}
		return AuditVerdict{}, fmt.Errorf("green video: http %d", code)
	}
	if resp.Body.Code == nil || *resp.Body.Code != 200 {
		code := int32(0)
		if resp.Body.Code != nil {
			code = *resp.Body.Code
		}
		msg := strings.TrimSpace(tea.StringValue(resp.Body.Message))
		if msg != "" {
			return AuditVerdict{}, fmt.Errorf("green video: code %d msg %s", code, msg)
		}
		return AuditVerdict{}, fmt.Errorf("green video: code %d", code)
	}
	return AuditVerdict{Pass: true}, nil
}

// EffectiveGreen 业务侧统一入口：配置关闭时用 noop（enabled=false 时不产生 Green 账单）。
func EffectiveGreen() GreenModerator {
	m := Green()
	if m.Enabled() {
		return m
	}
	return &noopGreenModerator{}
}
