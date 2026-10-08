// =================================================================================
package dao

import (
	"hello/internal/dao/internal"
)

type internalXiaozhiMcpBindingDao = *internal.XiaozhiMcpBindingDao

type xiaozhiMcpBindingDao struct {
	internalXiaozhiMcpBindingDao
}

var (
	// XiaozhiMcpBinding 小智 MCP 音箱绑定表。
	XiaozhiMcpBinding = xiaozhiMcpBindingDao{
		internal.NewXiaozhiMcpBindingDao(),
	}
)
