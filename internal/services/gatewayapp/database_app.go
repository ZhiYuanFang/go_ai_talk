package gatewayapp

import (
	"os"

	"hello/internal/platform/dbcfg"
)

// ApplyAppDatabaseLinkFromEnv 在进程启动最早阶段将 APP_DB_LINK 写入 gdb「app」分组。
func ApplyAppDatabaseLinkFromEnv() {
	dbcfg.ApplyGroupFromEnv("gateway-app", "app", "APP_DB_LINK", "GF_DATABASE_APP_LINK")
}

// AppDatabaseNameFromLink 从 GoFrame DSN 解析库名（如 ai_voice_app_test）。
func AppDatabaseNameFromLink(link string) string {
	return dbcfg.DatabaseNameFromLink(link)
}

// ResolvedAppDatabaseName 返回当前进程 app 库名（来自 APP_DB_LINK，经 MYSQL_TCP_HOST 解析后仍取库名段）。
func ResolvedAppDatabaseName() string {
	link := os.Getenv("APP_DB_LINK")
	if host := os.Getenv("MYSQL_TCP_HOST"); host != "" {
		link = dbcfg.RewriteLinkHost(link, host)
	}
	return dbcfg.DatabaseNameFromLink(link)
}
