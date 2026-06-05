package dbcfg

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/glog"
)

// linkHostRe 匹配 GoFrame DSN 中 @tcp(host:port) 的 host 段，便于 MYSQL_TCP_HOST 覆盖。
var linkHostRe = regexp.MustCompile(`(@tcp\()([^:/]+)(:\d+\))`)

// ApplyGroupFromEnv 在任意 g.DB(group) 之前，用 envKey 对应 DSN 写入 gdb 分组（Compose/.env 为唯一来源）。
// 须在设置 GF_GCFG_FILE 之后调用；manifest/config 中不再保留 database.*.link 占位。
func ApplyGroupFromEnv(service, group, envKey, gfEnvKey string) {
	link := resolveEffectiveLink(strings.TrimSpace(os.Getenv(envKey)))
	if link == "" {
		return
	}
	if gfEnvKey != "" {
		_ = os.Setenv(gfEnvKey, link)
	}

	node := gdb.ConfigNode{Link: link, Role: "master"}
	gdb.SetConfigGroup(group, gdb.ConfigGroup{node})
	glog.Printf(context.Background(), "[%s] database.%s 已用 %s 配置，库名=%s，主机=%s",
		service, group, envKey, DatabaseNameFromLink(link), HostFromLink(link))
}

// resolveEffectiveLink 应用 MYSQL_TCP_HOST 覆盖（若已设置）。
func resolveEffectiveLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	host := strings.TrimSpace(os.Getenv("MYSQL_TCP_HOST"))
	if host == "" {
		return link
	}
	return RewriteLinkHost(link, host)
}

// RewriteLinkHost 将 DSN 中 @tcp(host:port) 的 host 替换为 newHost。
func RewriteLinkHost(link, newHost string) string {
	link = strings.TrimSpace(link)
	newHost = strings.TrimSpace(newHost)
	if link == "" || newHost == "" {
		return link
	}
	return linkHostRe.ReplaceAllString(link, "${1}"+newHost+"${3}")
}

// HostFromLink 从 GoFrame DSN 解析 @tcp(host:port) 中的 host。
func HostFromLink(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	m := linkHostRe.FindStringSubmatch(link)
	if len(m) >= 3 {
		return m[2]
	}
	return ""
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
