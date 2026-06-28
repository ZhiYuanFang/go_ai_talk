package v1

import "github.com/gogf/gf/v2/frame/g"

// AppStatusAdminLoginReq 运维登录（复用 Hub 账号口令，签发 Admin JWT）。
type AppStatusAdminLoginReq struct {
	g.Meta   `path:"/admin/api/login" method:"post" tags:"app-status-admin" summary:"运维登录签发 Admin JWT"`
	Username string `json:"username" v:"required#username 不能为空"`
	Password string `json:"password" v:"required#password 不能为空"`
}

// AppStatusAdminLoginRes 登录成功返回 Admin JWT。
type AppStatusAdminLoginRes struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// AppStatusAdminBannerGetReq 读取当前 banner 配置。
type AppStatusAdminBannerGetReq struct {
	g.Meta `path:"/admin/api/banner" method:"get" tags:"app-status-admin" summary:"读取维护通知配置"`
}

// AppStatusAdminBannerGetRes 含 contentKey 供运维页展示内容指纹。
type AppStatusAdminBannerGetRes struct {
	Active        bool   `json:"active"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	ExpectedEndAt *int64 `json:"expectedEndAt,omitempty"`
	Dismissible   bool   `json:"dismissible"`
	UpdatedAt     int64  `json:"updatedAt"`
	ContentKey    string `json:"contentKey"`
}

// AppStatusAdminBannerPutReq 保存 banner；active=false 时 App 侧不再弹窗。
type AppStatusAdminBannerPutReq struct {
	g.Meta        `path:"/admin/api/banner" method:"put" tags:"app-status-admin" summary:"保存维护通知配置"`
	Active        bool   `json:"active"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	ExpectedEndAt *int64 `json:"expectedEndAt,omitempty"`
	Dismissible   bool   `json:"dismissible"`
}

// AppStatusAdminBannerPutRes 保存后的生效快照。
type AppStatusAdminBannerPutRes struct {
	Active        bool   `json:"active"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	ExpectedEndAt *int64 `json:"expectedEndAt,omitempty"`
	Dismissible   bool   `json:"dismissible"`
	UpdatedAt     int64  `json:"updatedAt"`
	ContentKey    string `json:"contentKey"`
}
