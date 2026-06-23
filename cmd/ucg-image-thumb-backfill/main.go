package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"hello/internal/platform/dbcfg"
	ucgsvc "hello/internal/services/ucg"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "仅列出将处理的 objectKey，不写 OSS")
	limit := flag.Int("limit", 0, "最多处理条数，0 表示不限制")
	concurrency := flag.Int("concurrency", 4, "并发 worker 数")
	envFile := flag.String("env-file", "manifest/docker/env/.env.test", "启动前加载的 dotenv 文件；空字符串表示不加载")
	flag.Parse()

	if strings.TrimSpace(*envFile) != "" {
		if err := loadEnvFile(*envFile); err != nil {
			fmt.Fprintf(os.Stderr, "load env file %s: %v\n", *envFile, err)
			os.Exit(1)
		}
	}
	prepareRuntime()
	ctx := gctx.New()

	keys, err := collectImageObjectKeys(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect keys failed: %v\n", err)
		os.Exit(1)
	}
	if *limit > 0 && len(keys) > *limit {
		keys = keys[:*limit]
	}

	var okN, missN, failN int64
	work := make(chan string, *concurrency)
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for key := range work {
			if *dryRun {
				fmt.Printf("[dry-run] would ensure thumb: %s\n", key)
				atomic.AddInt64(&okN, 1)
				continue
			}
			err := ucgsvc.EnsureImageThumb(ctx, key)
			if err != nil {
				msg := err.Error()
				if strings.Contains(msg, "原图不存在") {
					atomic.AddInt64(&missN, 1)
					fmt.Fprintf(os.Stderr, "[missing] %s: %v\n", key, err)
					continue
				}
				atomic.AddInt64(&failN, 1)
				fmt.Fprintf(os.Stderr, "[fail] %s: %v\n", key, err)
				continue
			}
			atomic.AddInt64(&okN, 1)
		}
	}

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker()
	}
	for _, key := range keys {
		work <- key
	}
	close(work)
	wg.Wait()

	fmt.Printf("done total=%d ok=%d missing_original=%d failed=%d dry_run=%v\n",
		len(keys), okN, missN, failN, *dryRun)
	if failN > 0 {
		os.Exit(1)
	}
}

func prepareRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.ucg-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("ucg-image-thumb-backfill", "default", "UCG_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
}

// loadEnvFile 解析 dotenv 并写入进程环境；已存在的变量不被覆盖（便于命令行 override）。
func loadEnvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		if key == "" {
			continue
		}
		if len(val) >= 2 {
			if val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			} else if val[0] == '\'' && val[len(val)-1] == '\'' {
				val = val[1 : len(val)-1]
			}
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return nil
}

func collectImageObjectKeys(ctx context.Context) ([]string, error) {
	const q = `
SELECT DISTINCT object_key AS k FROM ucg_media_blob WHERE media_kind = 1 AND object_key <> ''
UNION
SELECT DISTINCT object_key FROM ucg_post_media WHERE media_kind = 1 AND object_key <> ''
UNION
SELECT DISTINCT avatar_key FROM ucg_profile WHERE avatar_key <> ''
UNION
SELECT DISTINCT image_key FROM ucg_chat_message WHERE image_key <> ''`
	rows, err := g.DB().GetAll(ctx, q)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rows))
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row["k"].String())
		if key == "" || !ucgsvc.IsImageObjectKey(key) {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out, nil
}
