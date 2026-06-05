package controller

import (
	"context"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/model/entity"
	"hello/internal/services/contracts"
	"hello/internal/services/gatewayapp"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/glog"
)

// SiteHome GET /device/app/api/site/home，返回胖宝官网首页所需的公开聚合数据。
func (c *GatewayAppCtrl) SiteHome(ctx context.Context, req *v1.GatewayAppSiteHomeReq) (res *v1.GatewayAppSiteHomeRes, err error) {
	_ = req
	publicBaseURL := gatewayAppPublicBaseURL(ctx)
	events := fetchGatewayAppSiteEvents(ctx, publicBaseURL)
	android := buildGatewayAppSiteAndroidDownload(ctx, publicBaseURL)
	return &v1.GatewayAppSiteHomeRes{
		BrandName:      "胖宝",
		HeroTitle:      "专注母婴喂养服务，让照顾孩子更轻松",
		HeroSubtitle:   "围绕喂奶、睡眠、换尿布等关键喂养场景，帮助家庭更便捷、更轻松地照顾孩子。",
		ServiceSummary: "胖宝专注母婴喂养服务，通过更清晰的事件记录与下载体验，让日常照护更省心。",
		Events:         events,
		Android:        android,
		IOS: v1.GatewayAppSiteIOSDownload{
			SearchTerm:  "胖宝",
			Instruction: "前往 App Store 搜索“胖宝”下载",
		},
	}, nil
}

func fetchGatewayAppSiteEvents(ctx context.Context, publicBaseURL string) []v1.GatewayAppSiteEventItem {
	targets := contracts.ResolveHTTPTargets()
	url := strings.TrimSpace(targets.DeviceInternalEventOptionsURL())
	if url == "" {
		return []v1.GatewayAppSiteEventItem{}
	}
	resp, err := gclient.New().Get(ctx, url)
	if err != nil {
		glog.Warningf(ctx, "[gateway-app-site] 读取事件列表失败 err=%v", err)
		return []v1.GatewayAppSiteEventItem{}
	}
	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 {
		glog.Warningf(ctx, "[gateway-app-site] 事件列表返回失败 message=%s", j.Get("message").String())
		return []v1.GatewayAppSiteEventItem{}
	}
	list := j.GetJson("data.list")
	if list == nil {
		return []v1.GatewayAppSiteEventItem{}
	}
	rows := make([]entity.Event, 0)
	if err := list.Scan(&rows); err != nil {
		glog.Warningf(ctx, "[gateway-app-site] 解析事件列表失败 err=%v", err)
		return []v1.GatewayAppSiteEventItem{}
	}
	items := make([]v1.GatewayAppSiteEventItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, v1.GatewayAppSiteEventItem{
			Id:       row.Id,
			Name:     strings.TrimSpace(row.Name),
			LogoUrl:  gatewayAppAbsoluteAssetURL(publicBaseURL, strings.TrimSpace(row.Logo)),
			Color:    strings.TrimSpace(row.Color),
			ParentId: row.ParentId,
		})
	}
	return items
}

func buildGatewayAppSiteAndroidDownload(ctx context.Context, publicBaseURL string) v1.GatewayAppSiteAndroidDownload {
	row, ok := loadLatestAppVersionRow(ctx)
	if !ok {
		return v1.GatewayAppSiteAndroidDownload{
			Available:      false,
			ButtonText:     "Android 下载暂未开放",
			StatusText:     "Android 版本信息准备中",
			UnavailableTip: "当前暂无可用 Android 安装包，请稍后再试。",
		}
	}
	downloadURL := gatewayAppAbsoluteAssetURL(publicBaseURL, strings.TrimSpace(row.DownloadUrl))
	if downloadURL == "" {
		return v1.GatewayAppSiteAndroidDownload{
			Available:      false,
			LatestVersion:  strings.TrimSpace(row.LatestVersion),
			ReleaseNotes:   strings.TrimSpace(row.ReleaseNotes),
			ButtonText:     "Android 下载暂未开放",
			StatusText:     "Android 下载链接暂不可用",
			UnavailableTip: "当前暂无可用 Android 安装包，请稍后再试。",
		}
	}
	return v1.GatewayAppSiteAndroidDownload{
		Available:      true,
		LatestVersion:  strings.TrimSpace(row.LatestVersion),
		ReleaseNotes:   strings.TrimSpace(row.ReleaseNotes),
		DownloadUrl:    downloadURL,
		QrValue:        downloadURL,
		StatusText:     "扫码即可下载 Android 版本",
		ButtonText:     "下载 Android 版",
		UnavailableTip: "",
	}
}

func gatewayAppPublicBaseURL(ctx context.Context) string {
	// 优先使用部署配置的对外基址，确保 test/prod 二维码与下载链指向各自域名（不依赖反代 Host 头）。
	if configured := strings.TrimSpace(gatewayapp.PublicBaseURL(ctx)); configured != "" {
		return configured
	}
	r := ghttp.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	scheme := strings.TrimSpace(r.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = strings.TrimSpace(r.URL.Scheme)
	}
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = strings.TrimSpace(r.GetHeader("Host"))
	}
	if host == "" {
		return ""
	}
	return strings.TrimRight(scheme+"://"+host, "/")
}

func gatewayAppAbsoluteAssetURL(publicBaseURL, raw string) string {
	path := gatewayapp.NormalizeAssetPath(strings.TrimSpace(raw))
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if publicBaseURL == "" {
		return path
	}
	return strings.TrimRight(publicBaseURL, "/") + path
}
