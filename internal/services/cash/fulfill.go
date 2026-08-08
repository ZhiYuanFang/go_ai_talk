package cash

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// FulfillPaid 将订单标为 paid 并续期权益（幂等：已 paid 同 channel_txn 直接返回）。
func FulfillPaid(ctx context.Context, orderNo, channel, channelTxnID string, amountFen int) error {
	orderNo = strings.TrimSpace(orderNo)
	channel = strings.TrimSpace(channel)
	channelTxnID = strings.TrimSpace(channelTxnID)
	if orderNo == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "orderNo 不能为空")
	}

	// 渠道交易号已履约 → 幂等成功。
	if channelTxnID != "" {
		if existed, err := loadOrderByChannelTxn(ctx, channel, channelTxnID); err != nil {
			return err
		} else if existed != nil && existed.Status == OrderPaid {
			return nil
		}
	}

	order, err := loadOrderByNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if order.Channel != channel && channel != "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "支付渠道与订单不一致")
	}
	if amountFen > 0 && order.AmountFen != amountFen {
		return gerror.NewCode(gcode.CodeInvalidParameter, "支付金额与订单不一致")
	}
	if order.Status == OrderPaid {
		// 已支付：若本次带了新的 txn 且原为空则补写，不重复续期。
		if channelTxnID != "" && order.ChannelTxnId == "" {
			_, _ = g.DB().Model("vip_order").Ctx(ctx).Where("id", order.Id).Data(g.Map{
				"channel_txn_id": channelTxnID,
			}).Update()
		}
		return nil
	}
	if order.Status != OrderCreated {
		return gerror.NewCode(gcode.CodeInvalidOperation, "订单状态不可支付")
	}

	prod, err := GetActiveProduct(ctx, order.ProductCode)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("vip_order").Ctx(ctx).
		Where("id", order.Id).Where("status", OrderCreated).
		Data(g.Map{
			"status":         OrderPaid,
			"channel_txn_id": channelTxnID,
			"paid_at":        now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		// 并发下已被其它请求置 paid。
		return nil
	}
	expireAt, err := ExtendEntitlement(ctx, order.WxId, prod.DurationDays)
	if err != nil {
		glog.Errorf(ctx, "[cash] extend entitlement failed orderNo=%s wxId=%d err=%v", orderNo, order.WxId, err)
		return err
	}
	glog.Infof(ctx, "[cash] fulfilled orderNo=%s wxId=%d channel=%s expireAt=%d", orderNo, order.WxId, channel, expireAt)
	return nil
}
