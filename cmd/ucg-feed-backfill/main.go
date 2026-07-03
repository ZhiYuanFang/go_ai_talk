package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/cachekit"
	"hello/internal/platform/dbcfg"
	"hello/internal/platform/rediscfg"
	ucgsvc "hello/internal/services/ucg"
	_ "hello/internal/shared/runtime"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "仅统计将处理的行数，不写 Redis")
	pageSize := flag.Int("page-size", 200, "MySQL 分页大小")
	limit := flag.Int("limit", 0, "最多处理帖数，0 表示不限制")
	likesOnly := flag.Bool("likes-only", false, "仅重建 liked SET")
	postsOnly := flag.Bool("posts-only", false, "仅 backfill 帖 ZSET/GEO/snapshot")
	commentsOnly := flag.Bool("comments-only", false, "仅 backfill 已发布评论 Redis")
	votesOnly := flag.Bool("votes-only", false, "仅 backfill 辩论帖投票 Hash 与 snapshot 票数")
	envFile := flag.String("env-file", "manifest/docker/env/.env.prod", "启动前加载的 dotenv；空字符串跳过")
	flag.Parse()

	if strings.TrimSpace(*envFile) != "" {
		if err := loadEnvFile(*envFile); err != nil {
			fmt.Fprintf(os.Stderr, "load env file %s: %v\n", *envFile, err)
			os.Exit(1)
		}
	}
	prepareRuntime()
	ctx := gctx.New()
	if !*dryRun {
		ensureRedisReady(ctx)
	}

	var postOK, postFail, likeOK, likeFail, commentOK, commentFail, voteOK, voteFail int64
	if *votesOnly {
		voteOK, voteFail = backfillVotes(ctx, *dryRun, *pageSize, *limit)
	} else if *commentsOnly {
		commentOK, commentFail = backfillComments(ctx, *dryRun, *pageSize, *limit)
	} else {
		if !*likesOnly {
			postOK, postFail = backfillPosts(ctx, *dryRun, *pageSize, *limit)
		}
		if !*postsOnly {
			ok, fail := backfillLikes(ctx, *dryRun)
			likeOK, likeFail = ok, fail
		}
	}
	fmt.Printf("done posts_ok=%d posts_fail=%d likes_ok=%d likes_fail=%d comments_ok=%d comments_fail=%d votes_ok=%d votes_fail=%d dry_run=%v\n",
		postOK, postFail, likeOK, likeFail, commentOK, commentFail, voteOK, voteFail, *dryRun)
	if postFail > 0 || likeFail > 0 || commentFail > 0 || voteFail > 0 {
		os.Exit(1)
	}
}

func backfillPosts(ctx context.Context, dryRun bool, pageSize, limit int) (int64, int64) {
	var okN, failN int64
	var lastID uint64
	for {
		model := dao.UcgPost.Ctx(ctx).
			Where(dao.UcgPost.Columns().Status, ucgsvc.PostStatusPublished).
			WhereGT(dao.UcgPost.Columns().Id, lastID).
			OrderAsc(dao.UcgPost.Columns().Id).
			Limit(pageSize)
		rows, err := model.All()
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan posts failed: %v\n", err)
			return okN, failN + 1
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			var post entity.UcgPost
			if err = row.Struct(&post); err != nil {
				atomic.AddInt64(&failN, 1)
				continue
			}
			lastID = post.Id
			if dryRun {
				atomic.AddInt64(&okN, 1)
				continue
			}
			if err = ucgsvc.BackfillPublishedPostRedis(ctx, post.Id); err != nil {
				atomic.AddInt64(&failN, 1)
				fmt.Fprintf(os.Stderr, "[post-fail] id=%d err=%v\n", post.Id, err)
				continue
			}
			atomic.AddInt64(&okN, 1)
			if limit > 0 && okN+failN >= int64(limit) {
				return okN, failN
			}
		}
		if limit > 0 && okN+failN >= int64(limit) {
			break
		}
	}
	return okN, failN
}

func backfillComments(ctx context.Context, dryRun bool, pageSize, limit int) (int64, int64) {
	var okN, failN int64
	var lastID uint64
	for {
		model := dao.UcgPost.Ctx(ctx).
			Where(dao.UcgPost.Columns().Status, ucgsvc.PostStatusPublished).
			WhereGT(dao.UcgPost.Columns().CommentCount, 0).
			WhereGT(dao.UcgPost.Columns().Id, lastID).
			OrderAsc(dao.UcgPost.Columns().Id).
			Limit(pageSize)
		rows, err := model.All()
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan posts for comments failed: %v\n", err)
			return okN, failN + 1
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			var post entity.UcgPost
			if err = row.Struct(&post); err != nil {
				atomic.AddInt64(&failN, 1)
				continue
			}
			lastID = post.Id
			if dryRun {
				atomic.AddInt64(&okN, 1)
				continue
			}
			if err = ucgsvc.BackfillPostCommentsRedis(ctx, post.Id); err != nil {
				atomic.AddInt64(&failN, 1)
				fmt.Fprintf(os.Stderr, "[comment-fail] postId=%d err=%v\n", post.Id, err)
				continue
			}
			atomic.AddInt64(&okN, 1)
			if limit > 0 && okN+failN >= int64(limit) {
				return okN, failN
			}
		}
		if limit > 0 && okN+failN >= int64(limit) {
			break
		}
	}
	return okN, failN
}

func backfillVotes(ctx context.Context, dryRun bool, pageSize, limit int) (int64, int64) {
	var okN, failN int64
	var lastID uint64
	for {
		model := dao.UcgPost.Ctx(ctx).
			Where(dao.UcgPost.Columns().Status, ucgsvc.PostStatusPublished).
			Where(dao.UcgPost.Columns().Type, ucgsvc.PostTypeDebate).
			WhereGT(dao.UcgPost.Columns().Id, lastID).
			OrderAsc(dao.UcgPost.Columns().Id).
			Limit(pageSize)
		rows, err := model.All()
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan debate posts for votes failed: %v\n", err)
			return okN, failN + 1
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			var post entity.UcgPost
			if err = row.Struct(&post); err != nil {
				atomic.AddInt64(&failN, 1)
				continue
			}
			lastID = post.Id
			if dryRun {
				atomic.AddInt64(&okN, 1)
				continue
			}
			if err = ucgsvc.BackfillPostVoteCountsRedis(ctx, post.Id); err != nil {
				atomic.AddInt64(&failN, 1)
				fmt.Fprintf(os.Stderr, "[vote-fail] postId=%d err=%v\n", post.Id, err)
				continue
			}
			atomic.AddInt64(&okN, 1)
			if limit > 0 && okN+failN >= int64(limit) {
				return okN, failN
			}
		}
		if limit > 0 && okN+failN >= int64(limit) {
			break
		}
	}
	return okN, failN
}

func backfillLikes(ctx context.Context, dryRun bool) (int64, int64) {
	rows, err := g.DB().GetAll(ctx, `
SELECT wx_id, post_id FROM ucg_post_like ORDER BY wx_id, post_id`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan likes failed: %v\n", err)
		return 0, 1
	}
	type pair struct {
		wxID   int64
		postID uint64
	}
	byUser := make(map[int64][]uint64)
	for _, row := range rows {
		wxID := row["wx_id"].Int64()
		postID := row["post_id"].Uint64()
		if wxID <= 0 || postID == 0 {
			continue
		}
		byUser[wxID] = append(byUser[wxID], postID)
	}
	var okN, failN int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for wxID, ids := range byUser {
		wg.Add(1)
		sem <- struct{}{}
		go func(wxID int64, ids []uint64) {
			defer wg.Done()
			defer func() { <-sem }()
			if dryRun {
				atomic.AddInt64(&okN, int64(len(ids)))
				return
			}
			if err := ucgsvc.BackfillUserLikedPosts(ctx, wxID, ids); err != nil {
				atomic.AddInt64(&failN, 1)
				fmt.Fprintf(os.Stderr, "[like-fail] wxId=%d err=%v\n", wxID, err)
				return
			}
			atomic.AddInt64(&okN, int64(len(ids)))
		}(wxID, ids)
	}
	wg.Wait()
	return okN, failN
}

func prepareRuntime() {
	if strings.TrimSpace(os.Getenv("GF_GCFG_FILE")) == "" {
		_ = os.Setenv("GF_GCFG_FILE", "manifest/config/config.ucg-service.yaml")
	}
	dbcfg.ApplyGroupFromEnv("ucg-feed-backfill", "default", "UCG_DB_LINK", "GF_DATABASE_DEFAULT_LINK")
	// 与 ucg-service 一致：Redis 仅经 GF_REDIS_DEFAULT_ADDRESS 注入（yaml 无 redis 段）。
	rediscfg.ApplyDefaultFromEnv("ucg-feed-backfill")
}

// ensureRedisReady 非 dry-run 写 Redis 前校验配置与连通性，避免 g.Redis() panic。
func ensureRedisReady(ctx context.Context) {
	if rediscfg.DefaultAddressFromEnv() == "" {
		fmt.Fprintln(os.Stderr, "GF_REDIS_DEFAULT_ADDRESS 未配置：请在 env-file 中设置 Redis 地址。")
		fmt.Fprintln(os.Stderr, "宿主机执行时须使用可连通的地址（非 Docker 服务名）；或在 ucg-service 容器内运行本命令。")
		os.Exit(2)
	}
	if err := cachekit.Default().Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Redis 连通失败 address=%s err=%v\n", rediscfg.DefaultAddressFromEnv(), err)
		os.Exit(2)
	}
}

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
