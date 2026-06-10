package do

import "github.com/gogf/gf/v2/frame/g"

// AiQuotaUserOverride is the golang structure of table ai_quota_user_override for DAO operations.
type AiQuotaUserOverride struct {
	g.Meta              `orm:"table:ai_quota_user_override, do:true"`
	WxId                interface{}
	PolishMonthlyLimit  interface{}
	VoiceAiMonthlyLimit interface{}
	UpdatedAt           interface{}
}
