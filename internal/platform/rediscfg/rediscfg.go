package rediscfg

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/os/glog"
)

// DefaultAddressFromEnv 返回 Compose/.env 注入的 GF_REDIS_DEFAULT_ADDRESS。
// manifest/config 已移除 redis 段；Publish（g.Redis）与 Subscribe（redismsgkit）须共用此来源。
func DefaultAddressFromEnv() string {
	return strings.TrimSpace(os.Getenv("GF_REDIS_DEFAULT_ADDRESS"))
}

// ApplyDefaultFromEnv 在任意 g.Redis() 之前，用 GF_REDIS_DEFAULT_ADDRESS 写入 gredis default 分组。
// 多地址逗号分隔时 go-redis 走 Cluster 客户端；单地址为 standalone。
func ApplyDefaultFromEnv(service string) {
	addr := DefaultAddressFromEnv()
	if addr == "" {
		return
	}
	cfg := &gredis.Config{
		Address: addr,
		Db:      redisDBFromEnv(),
	}
	gredis.SetConfig(cfg)
	glog.Printf(context.Background(), "[%s] redis.default 已用 GF_REDIS_DEFAULT_ADDRESS 配置，address=%s db=%d",
		service, addr, cfg.Db)
}

func redisDBFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("GF_REDIS_DEFAULT_DB"))
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
