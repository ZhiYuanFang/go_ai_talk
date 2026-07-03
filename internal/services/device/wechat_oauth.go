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

const wechatOAuthAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
const wechatMiniProgramCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// wechatOAuthResult 微信开放平台 OAuth 换票成功时的业务字段（微信侧令牌不落库、不记录日志原文）。
type wechatOAuthResult struct {
	OpenID  string
	UnionID string
}

// exchangeAuthCodeForUnionID 使用服务端持有的 appId/appSecret 将授权临时 code 换为 openid/unionid。
// jsCode 为 HTTP 入参字段名，语义为微信开放平台授权 code（移动应用 SendAuth 或网站应用 qrconnect 回调）。
// 多端统一身份依赖 unionid；若主体未绑定微信开放平台，unionid 为空时返回错误。
func exchangeAuthCodeForUnionID(ctx context.Context, platform, jsCode string) (*wechatOAuthResult, error) {
	jsCode = strings.TrimSpace(jsCode)
	if jsCode == "" {
		return nil, errors.New("jsCode 不能为空")
	}
	platform = strings.TrimSpace(platform)
	if platform == "miniprogram" {
		return exchangeMiniProgramCodeForUnionID(ctx, platform, jsCode)
	}
	appID, secret, err := wechatPlatformCredentials(ctx, platform)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("appid", appID)
	q.Set("secret", secret)
	q.Set("code", jsCode)
	q.Set("grant_type", "authorization_code")
	full := wechatOAuthAccessTokenURL + "?" + q.Encode()

	client := gclient.New().Timeout(8 * time.Second)
	resp, err := client.Get(ctx, full)
	if err != nil {
		return nil, fmt.Errorf("调用微信 oauth2/access_token 失败: %w", err)
	}
	defer resp.Close()
	body := resp.ReadAll()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("微信 oauth2/access_token HTTP 状态异常: %d", resp.StatusCode)
	}

	var raw struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		OpenID       string `json:"openid"`
		Scope        string `json:"scope"`
		UnionID      string `json:"unionid"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析微信响应失败: %w", err)
	}
	if raw.ErrCode != 0 {
		// 常见：40029 code 无效；40163 code 已使用。不向客户端透传微信 errmsg 全文。
		glog.Warningf(ctx, "[wechat-oauth] access_token 业务失败 platform=%s errcode=%d", platform, raw.ErrCode)
		return nil, fmt.Errorf("微信登录校验失败（errcode=%d）", raw.ErrCode)
	}
	openid := strings.TrimSpace(raw.OpenID)
	if openid == "" {
		return nil, errors.New("微信未返回 openid")
	}
	unionid := strings.TrimSpace(raw.UnionID)
	if unionid == "" {
		return nil, errors.New("当前应用未返回 unionid，请确认已绑定微信开放平台且用户已授权")
	}
	// 明确丢弃微信 OAuth 令牌，登录链路仅需 unionid。
	_ = raw.AccessToken
	_ = raw.RefreshToken
	return &wechatOAuthResult{
		OpenID:  openid,
		UnionID: unionid,
	}, nil
}

// wechatPlatformCredentials 从 device 配置读取指定 platform 的 appId/appSecret。
func wechatPlatformCredentials(ctx context.Context, platform string) (appID, secret string, err error) {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return "", "", errors.New("platform 不能为空")
	}
	base := "wechat.platforms." + platform
	appID = strings.TrimSpace(g.Cfg().MustGet(ctx, base+".appId").String())
	secret = strings.TrimSpace(g.Cfg().MustGet(ctx, base+".appSecret").String())
	if appID == "" || secret == "" {
		return "", "", fmt.Errorf("未配置微信凭据 wechat.platforms.%s 的 appId/appSecret", platform)
	}
	return appID, secret, nil
}

// exchangeMiniProgramCodeForUnionID 小程序 wx.login code → jscode2session。
func exchangeMiniProgramCodeForUnionID(ctx context.Context, platform, jsCode string) (*wechatOAuthResult, error) {
	appID, secret, err := wechatPlatformCredentials(ctx, platform)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("appid", appID)
	q.Set("secret", secret)
	q.Set("js_code", jsCode)
	q.Set("grant_type", "authorization_code")
	full := wechatMiniProgramCode2SessionURL + "?" + q.Encode()

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
		return nil, fmt.Errorf("解析微信 jscode2session 响应失败: %w", err)
	}
	if raw.ErrCode != 0 {
		glog.Warningf(ctx, "[wechat-oauth] jscode2session 业务失败 platform=%s errcode=%d", platform, raw.ErrCode)
		return nil, fmt.Errorf("微信小程序登录校验失败（errcode=%d）", raw.ErrCode)
	}
	openid := strings.TrimSpace(raw.OpenID)
	if openid == "" {
		return nil, errors.New("微信未返回 openid")
	}
	unionid := strings.TrimSpace(raw.UnionID)
	if unionid == "" {
		return nil, errors.New("当前小程序未返回 unionid，请确认已绑定微信开放平台")
	}
	_ = raw.SessionKey
	return &wechatOAuthResult{
		OpenID:  openid,
		UnionID: unionid,
	}, nil
}
