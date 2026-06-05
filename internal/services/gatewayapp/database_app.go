package gatewayapp

import (
	"os"

	"hello/internal/platform/dbcfg"

	"github.com/gogf/gf/v2/database/gdb"
)

// ApplyAppDatabaseLinkFromEnv 在进程启动最早阶段将 APP_DB_LINK 写入 gdb「app」分组。
func ApplyAppDatabaseLinkFromEnv() {
	dbcfg.ApplyGroupFromEnv("gateway-app", "app", "APP_DB_LINK", "GF_DATABASE_APP_LINK")
}

// AppDatabaseNameFromLink 从 GoFrame DSN 解析库名（如 ai_voice_app_test）。
func AppDatabaseNameFromLink(link string) string {
	return dbcfg.DatabaseNameFromLink(link)
}

// ResolvedAppDatabaseName 返回当前进程 app 库名（优先 APP_DB_LINK）。
func ResolvedAppDatabaseName() string {
	if name := dbcfg.DatabaseNameFromLink(os.Getenv("APP_DB_LINK")); name != "" {
		return name
	}
	if cg := gdb.GetConfig("app"); len(cg) > 0 {
		return dbcfg.DatabaseNameFromLink(cg[0].Link)
	}
	return ""
}
