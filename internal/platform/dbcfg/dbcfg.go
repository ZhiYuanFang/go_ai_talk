package dbcfg

import (
	"context"
	"os"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/glog"
)

// ApplyGroupFromEnv 在任意 g.DB(group) 之前，用 envKey 对应 DSN 强制覆盖 yaml 中 database.{group}.link。
// 仅 Setenv("GF_DATABASE_*") 无效：yaml 已有 link 时 GoFrame 不会自动用环境变量替换。
func ApplyGroupFromEnv(service, group, envKey, gfEnvKey string) {
	link := strings.TrimSpace(os.Getenv(envKey))
	if link == "" {
		return
	}
	if gfEnvKey != "" {
		_ = os.Setenv(gfEnvKey, link)
	}
	gdb.SetConfigGroup(group, gdb.ConfigGroup{
		{Link: link, Role: "master"},
	})
	glog.Printf(context.Background(), "[%s] database.%s 已用 %s 覆盖，库名=%s",
		service, group, envKey, DatabaseNameFromLink(link))
}

// DatabaseNameFromLink 从 GoFrame DSN 解析库名（如 ai_voice_history_test）。
func DatabaseNameFromLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	if i := strings.LastIndex(link, "/"); i >= 0 && i < len(link)-1 {
		name := link[i+1:]
		if j := strings.Index(name, "?"); j >= 0 {
			name = name[:j]
		}
		return strings.TrimSpace(name)
	}
	return ""
}
