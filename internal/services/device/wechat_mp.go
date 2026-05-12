package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/glog"
)

const wechatCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// wechatCode2SessionResult 微信 jscode2session 成功时的业务字段（session_key 不落库、不记录日志原文）。
type wechatCode2SessionResult struct {
	OpenID     string
	UnionID    string
	SessionKey string
}

// exchangeJsCodeForUnionID 使用服务端持有的 appId/appSecret 将临时 js_code 换为 openid/unionid。
// 多端统一身份依赖 unionid；若主体未绑定微信开放平台，unionid 为空，此时返回错误提示接入方。
func exchangeJsCodeForUnionID(ctx context.Context, platform, jsCode string) (*wechatCode2SessionResult, error) {
	jsCode = strings.TrimSpace(jsCode)
	if jsCode == "" {
		return nil, errors.New("jsCode 不能为空")
	}
	appID, secret, err := wechatMiniProgramCredentials(ctx, platform)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("appid", appID)
	q.Set("secret", secret)
	q.Set("js_code", jsCode)
	q.Set("grant_type", "authorization_code")
	full := wechatCode2SessionURL + "?" + q.Encode()

	client := gclient.New().Timeout(8 * time.Second)
	resp, err := client.Get(ctx, full)
	if err != nil {
		return nil, fmt.Errorf("调用微信 jscode2session 失败: %w", err)
	}
	defer resp.Close()
	body := resp.ReadAll()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信 jscode2session HTTP 状态异常: %d", resp.StatusCode)
	}

	var raw struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if raw.ErrCode != 0 {
		// 常见：40029 code 无效；40226 高风险用户等。不向客户端透传微信原文，避免泄露内部判断。
		glog.Warningf(ctx, "[wechat-mp] jscode2session 业务失败 platform=%s errcode=%d", platform, raw.ErrCode)
		return nil, fmt.Errorf("微信登录校验失败（errcode=%d）", raw.ErrCode)
	}
	openid := strings.TrimSpace(raw.OpenID)
	if openid == "" {
		return nil, errors.New("微信未返回 openid")
	}
	unionid := strings.TrimSpace(raw.UnionID)
	if unionid == "" {
		return nil, errors.New("当前小程序未返回 unionid，请确认已绑定微信开放平台且用户已授权")
	}
	return &wechatCode2SessionResult{
		OpenID:     openid,
		UnionID:    unionid,
		SessionKey: strings.TrimSpace(raw.SessionKey),
	}, nil
}

// wechatMiniProgramCredentials 从 device 配置读取指定 platform 的 appId/appSecret。
func wechatMiniProgramCredentials(ctx context.Context, platform string) (appID, secret string, err error) {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return "", "", errors.New("platform 不能为空")
	}
	base := "wechatMp.platforms." + platform
	appID = strings.TrimSpace(g.Cfg().MustGet(ctx, base+".appId").String())
	secret = strings.TrimSpace(g.Cfg().MustGet(ctx, base+".appSecret").String())
	if appID == "" || secret == "" {
		return "", "", fmt.Errorf("未配置小程序凭据 wechatMp.platforms.%s 的 appId/appSecret", platform)
	}
	return appID, secret, nil
}
