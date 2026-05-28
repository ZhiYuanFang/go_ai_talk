package gatewayapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	// IOSBundleID 为胖宝 iOS 应用固定 Bundle ID。
	IOSBundleID = "com.fzy.pangbao"
	// universalLinksRoutePrefix 为微信开放平台与客户端统一使用的 Universal Links 前缀。
	universalLinksRoutePrefix = "/wx/ulink/"
	// universalLinksPathPattern 为 AASA 中放行的路径模式。
	universalLinksPathPattern = "/wx/ulink/*"
)

type appleAppSiteAssociation struct {
	Applinks appleAppSiteAssociationApplinks `json:"applinks"`
}

type appleAppSiteAssociationApplinks struct {
	Details []appleAppSiteAssociationDetail `json:"details"`
}

type appleAppSiteAssociationDetail struct {
	AppIDs     []string                          `json:"appIDs"`
	Components []appleAppSiteAssociationPathRule `json:"components"`
}

type appleAppSiteAssociationPathRule struct {
	Path    string `json:"/"`
	Comment string `json:"comment,omitempty"`
}

// IOSTeamID 返回当前环境配置的 Apple Team ID。
// 优先环境变量 GATEWAY_APP_IOS_TEAM_ID，其次读取 gatewayApp.ios.teamId。
func IOSTeamID(ctx context.Context) string {
	if v := strings.TrimSpace(os.Getenv("GATEWAY_APP_IOS_TEAM_ID")); v != "" {
		return v
	}
	return strings.TrimSpace(g.Cfg().MustGet(ctx, "gatewayApp.ios.teamId").String())
}

// UniversalLinksPathPrefix 返回微信开放平台应填写的统一路径前缀。
func UniversalLinksPathPrefix() string {
	return universalLinksRoutePrefix
}

// UniversalLinksPublicURL 返回面向微信开放平台和 iOS 客户端配置的完整 Universal Links 基址。
// 当 publicBaseUrl 配为 https://www.pangbao.cuplay.top 时，此处返回 https://www.pangbao.cuplay.top/wx/ulink/。
func UniversalLinksPublicURL(ctx context.Context) string {
	base := strings.TrimRight(PublicBaseURL(ctx), "/")
	if base == "" {
		return ""
	}
	return base + UniversalLinksPathPrefix()
}

// BuildAppleAppSiteAssociation 根据当前 Team ID 生成 AASA 文件内容。
// Team ID 未配置时返回显式错误，避免输出伪造 appIDs 误导 Apple/微信校验。
func BuildAppleAppSiteAssociation(ctx context.Context) ([]byte, error) {
	teamID := IOSTeamID(ctx)
	if teamID == "" {
		return nil, fmt.Errorf("gatewayApp.ios.teamId 未配置，无法生成 apple-app-site-association")
	}
	payload := appleAppSiteAssociation{
		Applinks: appleAppSiteAssociationApplinks{
			Details: []appleAppSiteAssociationDetail{
				{
					AppIDs: []string{teamID + "." + IOSBundleID},
					Components: []appleAppSiteAssociationPathRule{
						{
							Path:    universalLinksPathPattern,
							Comment: "微信开放平台 Universal Links 前缀",
						},
					},
				},
			},
		},
	}
	return json.Marshal(payload)
}
