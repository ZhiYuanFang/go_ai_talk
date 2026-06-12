package ucg

import (
	"context"

	"hello/internal/platform/eventkit"

	"github.com/gogf/gf/v2/os/glog"
)

// StartUcgMQConsumers 启动 UCG 审核/推荐 AMQP push consumer（共用单 connection）。
func StartUcgMQConsumers(ctx context.Context) {
	if !ucgAuditConsumerEnabled() {
		glog.Infof(ctx, "[ucg-audit-mq] consumer disabled by %s", ucgAuditConsumerEnabledEnv)
	}

	if ucgRecommendConsumerEnabled() {
		StartRecommendHotReconciler(ctx)
	} else {
		glog.Infof(ctx, "[ucg-recommend-mq] consumer disabled by %s", ucgRecommendConsumerEnabledEnv)
	}

	if !ucgAuditConsumerEnabled() && !ucgRecommendConsumerEnabled() {
		return
	}

	amqpURL, err := ucgAuditAMQPURL()
	if err != nil {
		glog.Warningf(ctx, "[ucg-mq] AMQP disabled: %v", err)
		return
	}

	var subs []eventkit.AMQPQueueSubscription
	if ucgAuditConsumerEnabled() {
		prefetch := ucgAuditPrefetch()
		for _, q := range ucgAuditQueues {
			subs = append(subs, eventkit.AMQPQueueSubscription{
				QueueName:         q,
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
