package ucg

import (
	"context"

	"hello/internal/platform/eventkit"
	"hello/internal/shared/mq"

	"github.com/gogf/gf/v2/os/glog"
)

// UcgRecommendPublisher 发布 UCG 推荐分增量事件；失败仅 log，不阻断主路径。
type UcgRecommendPublisher struct {
	pub eventkit.Publisher
}

func NewUcgRecommendPublisher() (*UcgRecommendPublisher, error) {
	pub, err := mq.NewObservedEventPublisher()
	if err != nil {
		return nil, err
	}
	return &UcgRecommendPublisher{pub: pub}, nil
}

func defaultUcgRecommendPublisher(ctx context.Context) *UcgRecommendPublisher {
	p, err := NewUcgRecommendPublisher()
	if err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] publisher init failed: %v", err)
		return nil
	}
	return p
}

func (p *UcgRecommendPublisher) publish(ctx context.Context, key eventkit.RouteKey, payload map[string]any) error {
	if p == nil || p.pub == nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] publish skipped (no publisher) key=%s payload=%v", key, payload)
		return eventkit.ErrUnavailable
	}
	return p.pub.Publish(ctx, key.String(), payload)
}

// PublishPostPublished 帖过审发布。
func PublishPostPublished(ctx context.Context, postID uint64) {
	p := defaultUcgRecommendPublisher(ctx)
	if err := p.publish(ctx, eventkit.RoutingUcgPostPublished, map[string]any{"postId": postID}); err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] post.published publish failed postId=%d err=%v", postID, err)
	}
}

// PublishPostUnpublished 删帖/下架（原 published）。
func PublishPostUnpublished(ctx context.Context, postID uint64) {
	p := defaultUcgRecommendPublisher(ctx)
	if err := p.publish(ctx, eventkit.RoutingUcgPostUnpublished, map[string]any{"postId": postID}); err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] post.unpublished publish failed postId=%d err=%v", postID, err)
	}
}

// PublishPostLiked 点赞。
func PublishPostLiked(ctx context.Context, postID uint64) {
	p := defaultUcgRecommendPublisher(ctx)
	if err := p.publish(ctx, eventkit.RoutingUcgPostLiked, map[string]any{"postId": postID}); err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] post.liked publish failed postId=%d err=%v", postID, err)
	}
}

// PublishPostUnliked 取消赞。
func PublishPostUnliked(ctx context.Context, postID uint64) {
	p := defaultUcgRecommendPublisher(ctx)
	if err := p.publish(ctx, eventkit.RoutingUcgPostUnliked, map[string]any{"postId": postID}); err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] post.unliked publish failed postId=%d err=%v", postID, err)
	}
}

// PublishCommentPublished 评论过审。
func PublishCommentPublished(ctx context.Context, postID, commentID uint64) {
	p := defaultUcgRecommendPublisher(ctx)
	if err := p.publish(ctx, eventkit.RoutingUcgCommentPublished, map[string]any{
		"postId":    postID,
		"commentId": commentID,
	}); err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] comment.published publish failed postId=%d commentId=%d err=%v",
			postID, commentID, err)
	}
}

// PublishCommentRemoved 评论删除（原 published）。
func PublishCommentRemoved(ctx context.Context, postID, commentID uint64) {
	p := defaultUcgRecommendPublisher(ctx)
	if err := p.publish(ctx, eventkit.RoutingUcgCommentRemoved, map[string]any{
		"postId":    postID,
		"commentId": commentID,
	}); err != nil {
		glog.Warningf(ctx, "[ucg-recommend-mq] comment.removed publish failed postId=%d commentId=%d err=%v",
			postID, commentID, err)
	}
}
