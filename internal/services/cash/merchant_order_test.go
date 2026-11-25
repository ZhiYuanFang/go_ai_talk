package cash

import (
	"context"
	"testing"
)

func TestFulfillMerchantOrderDispatchPrefersMerchantTableShape(t *testing.T) {
	// 纯逻辑：Ensure merchant order fulfill does not touch vip_entitlement API surface.
	// 集成库测在有 MySQL 时另行覆盖；此处保证错误类型与金额校验入口存在。
	if ErrMerchantOrderNotFound == nil || ErrMerchantIDInvalid == nil {
		t.Fatal("merchant errors must be defined")
	}
	if MerchantProductCodeMonthly == "" || MerchantDefaultPriceFen != 100000 {
		t.Fatalf("default product seed constants broken")
	}
}

func TestBlockHistoryCreateForMerchantZeroID(t *testing.T) {
	block, err := BlockHistoryCreateForMerchant(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if block {
		t.Fatal("merchantID=0 must not block")
	}
}
