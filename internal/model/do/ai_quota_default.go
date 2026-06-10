package do

import "github.com/gogf/gf/v2/frame/g"

// AiQuotaDefault is the golang structure of table ai_quota_default for DAO operations.
type AiQuotaDefault struct {
	g.Meta              `orm:"table:ai_quota_default, do:true"`
	Id                  interface{}
	PolishMonthlyLimit  interface{}
	VoiceAiMonthlyLimit interface{}
	UpdatedAt           interface{}
}
