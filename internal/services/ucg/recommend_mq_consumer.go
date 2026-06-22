package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hello/internal/platform/cachekit"
	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/os/glog"
)

const ucgRecommendScoreQueue = "ucg.recommend.score.q"

var recommendThrottleCache = cachekit.Default()

// ucgRecommendAMQPHandler 推荐分 MQ consumer：throttle 限频；unpublished 永远 Ack。
func ucgRecommendAMQPHandler(ctx context.Context, queueName, routingKey string, body []byte) error {
	_ = queueName
	key := eventkit.RouteKey(strings.TrimSpace(routingKey))
	switch key {
	case eventkit.RoutingUcgPostUnpublished:
		return handleRecommendUnpublished(ctx, body)
	case eventkit.RoutingUcgPostPublished:
		// 新帖 score 由热区 reconciler 写入；遗留 published 消息直接 Ack 跳过。
		return nil
	case eventkit.RoutingUcgPostLiked, eventkit.RoutingUcgPostUnliked,
		eventkit.RoutingUcgCommentPublished, eventkit.RoutingUcgCommentRemoved:
		return handleRecommendRecompute(ctx, body, true)
	default:
		glog.Warningf(ctx, "[ucg-recommend-mq] unknown routing key=%s body=%s", routingKey, string(body))
		return nil
	}
}

// handleRecommendUnpublished 处理未发布帖子事件。
func handleRecommendUnpublished(ctx context.Context, body []byte) error {
	postID, err := parseRecommendPostID(body)
	if err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] unpublished invalid payload err=%v body=%s", err, string(body))
		return nil
	}
	if postID == 0 {
		return nil
	}
	if err = RemoveRecommendScore(ctx, postID); err != nil {
		// 下架语义：DB 异常也 Ack，避免阻塞队列；Feed JOIN published 兜底。
		glog.Errorf(ctx, "[ucg-recommend-mq] unpublished delete failed postId=%d err=%v", postID, err)
	}
	return nil
}

// handleRecommendRecompute 处理帖子互动事件。
func handleRecommendRecompute(ctx context.Context, body []byte, throttle bool) error {
	postID, err := parseRecommendPostID(body)
	if err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] invalid payload err=%v body=%s", err, string(body))
		return nil
	}
	if postID == 0 {
		return nil
	}
	if throttle {
		hot, hotErr := isPostInRecommendHotZone(ctx, postID)
		if hotErr != nil {
			return hotErr
		}
		if hot {
			// 热区互动由 reconciler 批量收敛；冷区保留 MQ 重算以支持翻红。
			return nil
		}
		if !tryRecommendThrottle(ctx, postID) {
			return nil
		}
	}
	if err = RecomputeRecommendScore(ctx, postID); err != nil {
		return err
	}
	return nil
}

func parseRecommendPostID(body []byte) (uint64, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, err
	}
	id := uint64(jsonInt(raw, "postId"))
	if id == 0 {
		return 0, fmt.Errorf("missing postId")
	}
	return id, nil
}

// tryRecommendThrottle 单 key 500ms 窗口；Redis 不可用时放行重算。
func tryRecommendThrottle(ctx context.Context, postID uint64) bool {
	cfg := LoadRecommendConfig(ctx)
	ttl := time.Duration(cfg.LikeThrottleMs) * time.Millisecond
	key := cachekit.UCGRecommendThrottleKey(postID)
	ok, err := recommendThrottleCache.SetNXEX(ctx, key, "1", ttl)
	if err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] throttle redis err postId=%d err=%v", postID, err)
		return true
	}
	return ok
}
