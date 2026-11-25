package cash

import "errors"

var (
	errMerchantIDInvalid       = errors.New("merchantId 无效")
	errMerchantExpireAtInvalid = errors.New("expireAt 必须晚于当前时间")
	ErrMerchantIDInvalid       = errMerchantIDInvalid
	ErrMerchantExpireAtInvalid = errMerchantExpireAtInvalid
	ErrMerchantProductMissing  = errors.New("商户订阅商品不存在或未上架")
	ErrMerchantOrderNotFound   = errors.New("商户订单不存在")
	ErrMerchantPriceInvalid    = errors.New("商户月费必须为正整数分")
)
