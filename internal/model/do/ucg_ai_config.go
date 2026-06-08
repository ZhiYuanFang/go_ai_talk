// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UcgAiConfig is the golang structure of table ucg_ai_config for DAO operations like Where/Data.
type UcgAiConfig struct {
	g.Meta              `orm:"table:ucg_ai_config, do:true"`
	Id                  interface{} // singleton row id=1
	VisionModel         interface{} //
	MaxImagesPerRequest interface{} //
	UpdatedAt           interface{} // unix seconds
	UpdatedBy           interface{} // admin operator label
}
