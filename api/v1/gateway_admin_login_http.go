package v1

import "github.com/gogf/gf/v2/frame/g"

// GatewayAdminLoginReq Hub 运维登录。
type GatewayAdminLoginReq struct {
	g.Meta   `path:"/device/admin/api/login" method:"post" tags:"admin" summary:"运维 Hub 登录签发 Admin JWT"`
	Username string `json:"username" v:"required#username 不能为空"`
	Password string `json:"password" v:"required#password 不能为空"`
}

// GatewayAdminLoginRes 登录成功返回 Admin JWT。
type GatewayAdminLoginRes struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}
