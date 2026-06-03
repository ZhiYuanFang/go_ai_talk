package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green "github.com/alibabacloud-go/green-20220302/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

const rejectReasonDefault = "违规已下架"

// AuditVerdict Green 审核结论。
type AuditVerdict struct {
	Pass   bool
	Reason string
}

// GreenModerator 内容审核接口（帖子/资料）。
type GreenModerator interface {
	ModerateText(ctx context.Context, service, content string) (AuditVerdict, error)
	ModerateImageURL(ctx context.Context, imageURL string) (AuditVerdict, error)
	ModerateVideoURL(ctx context.Context, videoURL string) (AuditVerdict, error)
	Enabled() bool
}

type greenModerator struct {
	client  *green.Client
	enabled bool
}

var (
	greenOnce sync.Once
	greenIns  GreenModerator
)

// Green 返回 Green 审核客户端单例。
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
	return AuditVerdict{Pass: true}, nil
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
		return AuditVerdict{}, err
	}
	return parseTextModeration(resp)
}

func (g *greenModerator) ModerateImageURL(ctx context.Context, imageURL string) (AuditVerdict, error) {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return AuditVerdict{Pass: true}, nil
	}
	params, err := json.Marshal(map[string]interface{}{
		"imageUrl": imageURL,
		"dataId":   imageURL,
	})
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
	params, err := json.Marshal(map[string]interface{}{
		"url":    videoURL,
		"dataId": videoURL,
	})
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
		return AuditVerdict{}, fmt.Errorf("green text: code %d", code)
	}
	labels := strings.TrimSpace(tea.StringValue(resp.Body.Data.Labels))
	reason := strings.TrimSpace(tea.StringValue(resp.Body.Data.Reason))
	if labels == "" || labels == "nonLabel" {
		return AuditVerdict{Pass: true}, nil
	}
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
		return AuditVerdict{}, fmt.Errorf("green image: http status not ok")
	}
	if resp.Body.Code == nil || *resp.Body.Code != 200 {
		return AuditVerdict{}, fmt.Errorf("green image: code not ok")
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
		return AuditVerdict{}, fmt.Errorf("green video: http status not ok")
	}
	if resp.Body.Code == nil || *resp.Body.Code != 200 {
		return AuditVerdict{}, fmt.Errorf("green video: code not ok")
	}
	// 视频审核多为异步任务；MVP 在 API 返回 200 且无即时 risk 信息时视为通过，后续可接 task 轮询。
	return AuditVerdict{Pass: true}, nil
}

// EffectiveGreen 返回实际审核器：配置关闭时自动 pass（便于联调）。
func EffectiveGreen() GreenModerator {
	m := Green()
	if m.Enabled() {
		return m
	}
	return &noopGreenModerator{}
}
