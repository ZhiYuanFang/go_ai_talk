package ucg

import (
	"context"

	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/os/glog"
)

// StartUcgMQConsumers ucg-service 启动时注册审核+推荐 AMQP consumer（共用一条 connection）。
// 审核 consumer 关闭后队列消息仍积压，重启 ucg 不会 purge 队列。
func StartUcgMQConsumers(ctx context.Context) {
	// 如果审核 consumer 关闭，则不订阅审核队列，消息 ready 堆积但不调 Green
	if !ucgAuditConsumerEnabled() {
		// UCG_AUDIT_MQ_CONSUMER_ENABLED=false：不订阅审核队列，消息 ready 堆积但不调 Green
		glog.Infof(ctx, "[ucg-audit-mq] consumer disabled by %s", ucgAuditConsumerEnabledEnv)
	}

	// 启动推荐热区 reconciler
	if ucgRecommendConsumerEnabled() {
		StartRecommendHotReconciler(ctx)
	} else {
		glog.Infof(ctx, "[ucg-recommend-mq] consumer disabled by %s", ucgRecommendConsumerEnabledEnv)
	}

	if !ucgAuditConsumerEnabled() && !ucgRecommendConsumerEnabled() {
		return // 两个都关则完全不连 RabbitMQ
	}

	// 获取 AMQP URL
	amqpURL, err := ucgAuditAMQPURL()
	if err != nil {
		glog.Warningf(ctx, "[ucg-mq] AMQP disabled: %v", err)
		return
	}

	var subs []eventkit.AMQPQueueSubscription
	if ucgAuditConsumerEnabled() {
		// 获取审核队列 prefetch
		prefetch := ucgAuditPrefetch()
		for _, q := range ucgAuditQueues {
			// 每个审核队列一个 goroutine + channel；handler 均为 ucgAuditAMQPHandler
			subs = append(subs, eventkit.AMQPQueueSubscription{
				QueueName:         q, // 含 ucg.profile.patch.submitted.q
				Prefetch:          prefetch,
				ConsumerTagPrefix: "ucg-audit",
				Handler:           ucgAuditAMQPHandler,
			})
		}
	}
	if ucgRecommendConsumerEnabled() {
		subs = append(subs, eventkit.AMQPQueueSubscription{
			QueueName:         ucgRecommendScoreQueue,
			Prefetch:          ucgRecommendPrefetch(),
			ConsumerTagPrefix: "ucg-recommend",
			Handler:           ucgRecommendAMQPHandler,
		})
	}

	// RunSharedAMQPConsumers 阻塞至 ctx 取消；每条 delivery 走 eventkit.handleDelivery
	eventkit.RunSharedAMQPConsumers(ctx, eventkit.SharedAMQPConfig{
		URL:           amqpURL,
		Subscriptions: subs,
	})
	glog.Infof(ctx, "[ucg-mq] shared AMQP started subscriptions=%d url=%s", len(subs), redactAMQPURL(amqpURL))
}

// StartUcgAuditMQConsumer 兼容旧入口，委托 StartUcgMQConsumers。
func StartUcgAuditMQConsumer(ctx context.Context) {
	StartUcgMQConsumers(ctx)
}
