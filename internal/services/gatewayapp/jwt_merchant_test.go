package gatewayapp

import (
	"context"
	"os"
	"testing"
)

func TestSignAndParseMerchantAccess(t *testing.T) {
	_ = os.Setenv("GATEWAY_APP_JWT_SECRET", "test-merchant-jwt-secret-32bytes!!")
	ctx := context.Background()
	tok, err := SignMerchantAccess(ctx, 42)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	id, err := ParseMerchantAccessClaims(ctx, tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if id != 42 {
		t.Fatalf("want 42 got %d", id)
	}
	if PeekIsMerchantJWT(ctx, tok) != true {
		t.Fatal("peek should be true")
	}
}

func TestIsMerchantPath(t *testing.T) {
	if !IsMerchantPath("/device/merchant/api/baby/list") {
		t.Fatal("expected merchant path")
	}
	if IsMerchantPath("/device/app/api/user/detail") {
		t.Fatal("app path should not be merchant")
	}
}
